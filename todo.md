# todo

- First create DefUsagesLookup during the TypeGathering phase, then pass it on to TypeChecker
(so that we can use it w/ fields & classes)
- Implement identifier resolution for fields and classes
- Add usages, not just defs
- Tie in go to definition/find references
- Implement `.` binary op
- Implement resolution for static methods/fields (like `System.out.println()`)
- Automatically parse all files in project on startup
- Handle array types
- Fix squigglies? They're kind of in weird spots
- Add warnings for unused variable
- Change the LSP to stop using 1-based line numbers