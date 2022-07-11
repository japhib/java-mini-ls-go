package server

import (
	"context"
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/pkg/errors"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"java-mini-ls-go/parse/typecheck"
	"java-mini-ls-go/util"
	"time"
)

func (j *JavaLS) rescanEverything(ctx context.Context) {
	rescanCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	folders, err := j.client.WorkspaceFolders(rescanCtx)
	if err != nil {
		j.log.Error(fmt.Sprintf("error getting workspace folders for rescanEverything: %s", err.Error()))
		return
	}

	for _, folder := range folders {
		// TODO scan workspace folders in parallel
		err = j.rescanWorkspaceFolder(folder.URI)
		if err != nil {
			j.log.Error(fmt.Sprintf("error scanning workspace folder. Folder=`%s` Error=`%s`", folder.URI, err.Error()))
		}
	}
}

type textDocParsed struct {
	doc    protocol.TextDocumentItem
	parsed antlr.Tree
}

func (j *JavaLS) rescanWorkspaceFolder(folderURI string) error {
	folderPath, err := j.fileResolver.FileURIToPath(folderURI)
	if err != nil {
		return errors.Wrapf(err, "error converting file URI %s to path", folderURI)
	}

	allFiles, err := j.fileResolver.ListJavaFilesRecursive(folderPath)
	if err != nil {
		return errors.Wrapf(err, "error scanning path %s for files", folderPath)
	}

	// read files & create TextDocumentItems
	textDocuments := util.MapAsync(allFiles, func(filePath string) protocol.TextDocumentItem {
		return protocol.TextDocumentItem{
			URI:        uri.New(filePath),
			LanguageID: "java",
			// TODO: make sure version 0 is okay to use
			Version: 0,
			Text:    j.fileResolver.ReadFile(filePath),
		}
	})

	// parse all files
	tdsParsed := util.MapAsync(textDocuments, func(td protocol.TextDocumentItem) textDocParsed {
		return textDocParsed{
			doc:    td,
			parsed: j.parseTextDocument(td),
		}
	})

	// create defUsages for each file
	defUsagesMap := map[string]*typecheck.DefinitionsUsagesLookup{}
	for _, td := range tdsParsed {
		defUsagesMap[string(td.doc.URI)] = typecheck.NewDefinitionsUsagesLookup()
	}

	// Perform each pass on all files before proceeding
	util.EachAsync(tdsParsed, func(tdParsed textDocParsed) {
		typecheck.GatherTypesFirstPass(
			string(tdParsed.doc.URI),
			int(tdParsed.doc.Version),
			tdParsed.parsed,
			j.builtinTypes,
			j.userTypes,
			defUsagesMap[string(tdParsed.doc.URI)],
		)
	})
	util.EachAsync(tdsParsed, func(tdParsed textDocParsed) {
		typecheck.GatherTypesSecondPass(
			string(tdParsed.doc.URI),
			int(tdParsed.doc.Version),
			tdParsed.parsed,
			j.builtinTypes,
			j.userTypes,
			defUsagesMap[string(tdParsed.doc.URI)],
		)
	})
	util.EachAsync(tdsParsed, func(tdParsed textDocParsed) {
		typeCheckingResult := typecheck.CheckTypes(
			j.log,
			string(tdParsed.doc.URI),
			int(tdParsed.doc.Version),
			tdParsed.parsed,
			j.builtinTypes,
			j.userTypes,
			defUsagesMap[string(tdParsed.doc.URI)],
		)
		j.handleTypeCheckResult(tdParsed.doc, typeCheckingResult)
	})

	return nil
}
