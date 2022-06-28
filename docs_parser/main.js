
// TODO:
// - Handle array and varargs params
// - Use full package names in references to other types

const { parse, NodeType } = require('node-html-parser');
const axios = require('axios');
const path = require('path');
const fs = require('fs');
const fsPromises = fs.promises;
const util = require('util');
const { URL } = require('url');

const baseUrl = 'https://docs.oracle.com/en/java/javase/17/docs/api';
const cacheFolder = 'java_sdk_cache';
const outFilePath = path.join(__dirname, '..', 'java_stdlib.json');

function toCachedPath(url) {
  if (!url.startsWith(baseUrl)) {
    throw new Error(`url doesn't start with baseUrl: ${url}`);
  }

  const trimmed = url.substring(baseUrl.length);
  return path.join(__dirname, cacheFolder, trimmed);
}

function stripWhitespace(text) {
  // the html javaparser puts a bunch of unnecessary newlines in.
  // '\xa0' is non-breaking space.
  // FYI `replaceAll` only exists in Node v16+.
  return text.replaceAll(/[\s\xa0]+/g, ' ').trim();
}

const openingAngleBracket = '<'.charCodeAt(0);
const closingAngleBracket = '>'.charCodeAt(0);

function stripGenerics(text) {
  if (!text.includes('<')) {
    return text;
  }

  // We have to iterate through it point by point since the generics can
  // be fairly complicated, e.g. `addAll<A>(Collection<List<A, B>, C> c)`

  let nestingLevel = 0;
  const builtString = [];
  for (let i = 0; i < text.length; i++) {
    const ch = text.charCodeAt(i);

    if (ch === openingAngleBracket) {
      nestingLevel++;
    } if (ch === closingAngleBracket) {
      nestingLevel--;
    } else if (nestingLevel === 0) {
      builtString.push(ch);
    }
  }

  if (nestingLevel !== 0) {
    // Angle brackets didn't match up, so it probably wasn't actual generics.
    console.log(`Angle brackets didn't match up: ${text}`);
    return text;
  }

  return String.fromCharCode(...builtString);
}

async function getData(url) {
  const cachedPath = toCachedPath(url);
  if (fs.existsSync(cachedPath)) {
    return (await fsPromises.readFile(cachedPath)).toString();
  }

  const result = await axios.get(url);
  if (result.status !== 200) {
    throw new Error(`Non-200 status code: ${result.status}`)
  }

  // write to cache
  console.log(`Caching contents of ${url}`)
  await fsPromises.mkdir(path.dirname(cachedPath), { recursive: true });
  await fsPromises.writeFile(cachedPath, result.data);

  return result.data;
}

async function genType(url) {
  const htmlStr = await getData(url);
  const root = parse(htmlStr);

  const data = {};

  const titleStr = root.querySelector('h1.title').textContent;
  const titleSplit = titleStr.split(' ');
  data.type = titleSplit[0].toLowerCase();
  data.name = titleSplit[1];

  const subtitles = root.querySelectorAll('.header > .sub-title');
  for (const st of subtitles) {
    const label = st.querySelector('span').textContent;
    const content = st.querySelector('a').textContent;

    switch (label.toLowerCase()) {
      case 'package':
        data.package = content;
        break;
      case 'module':
        data.module = content;
        break;
    }
  }

  console.log(`getting type data for ${data.type} ${data.package}.${data.name}`);

  let extendsImplements = root.querySelector('.type-signature .extends-implements');
  if (extendsImplements) {
    extendsImplements = stripWhitespace(extendsImplements.textContent);
    const matches = /extends (.*?)( implements (.*))?$/.exec(extendsImplements);
    if (matches) {
      const extendsStr = matches[1];
      if (extendsStr) {
        // Interfaces can extend more than one other type, so this is a type list
        // instead of a single type
        data.extends = parseTypeList(matches[1]);
      }

      const implementsStr = matches[3];
      if (implementsStr) {
        data.implements = parseTypeList(implementsStr);
      }
    } else {
      throw new Error(`Can't parse extends-implements section, unexpected format: '${extendsImplements}'`)
    }
  }

  data.fields = parseTable('field', root, '.field-summary .summary-table');
  data.constructors = parseTable('constructor', root, '.constructor-summary .summary-table').map(parseArgs);

  data.methods = parseTable('method', root, '.method-summary .summary-table').map(parseArgs);
  if (!data.methods || data.methods.length === 0) {
    data.methods = parseTable('method', root, '#method-summary-table .summary-table').map(parseArgs);
  }

  // console.log(util.inspect(data, {depth: 99}));

  return data;
}

function parseTypeList(typeList) {
  return typeList.split(',').map(s => s.trim()).map(stripGenerics);
}

function parseTable(type, root, tableSelector) {
  let columns = ['modifierAndType', 'name', 'description'];;
  if (type === 'constructor') {
    columns = ['modifier', 'name', 'description'];
  }

  const table = root.querySelector(tableSelector);
  if (!table) {
    return [];
  }

  let numColumns = 3;
  if (table.classNames.includes('two-column-summary')) {
    numColumns = 2;

    if (type === 'constructor') {
      columns = ['name', 'description'];
    } else {
      console.error('ERROR: Non-constructor field has 2 columns!', type, table.classNames);
      return [];
    }
  }

  const tableChildren = root.querySelectorAll(tableSelector + ' > div');
  const items = [];
  for (let i = numColumns; i < tableChildren.length - (numColumns - 1); i += numColumns) {
    const item = {};

    for (let j = 0; j < numColumns; j++) {
      const columnName = columns[j];
      let content = stripWhitespace(stripGenerics(tableChildren[i + j].textContent));

      if (columnName === 'modifierAndType') {
        // modifierAndType could look like "static final String", in which case we want
        // modifiers to be ["static", "final"] and type to be "string"
        const modifiers = content.split(' ');
        const [ type ] = modifiers.splice(-1);
        item.modifiers = modifiers;
        item.type = type; 
      } else if (columnName === 'modifier') {
        item.modifiers = [content];
      } else {
        item[columnName] = content;
      }
    }

    items.push(item);
  }
  
  return items;
}

function parseArgs(item) {
  let { name } = item;
  name = stripGenerics(name);

  const matches = /(.*)\((.*)\)/.exec(name);
  if (!matches) {
    console.error(`ERROR: Can't parse args, method formatted wrong: ${util.inspect(item, { depth: 99 })}`);
    return item;
  }

  const baseName = matches[1];
  let args = matches[2];

  if (args === '') {
    args = [];
  } else {
    args = args.split(',')
      .map(a => {
        const split = a.trim().split(' ');
        return {
          type: split[0],
          name: split[1]
        };
      });
  }

  item.name = baseName;
  item.args = args;

  return item;
}

async function getModuleLinks(basePath) {
  const modulesHtml = await getData(basePath);
  const modulesRoot = parse(modulesHtml);

  return modulesRoot
    .querySelectorAll('#all-modules-table .summary-table .col-first a')
    .filter(a => a.textContent.trim().startsWith('java.'))
    .map(el => el.getAttribute('href'))
    .map(link => new URL(link, basePath).toString());
}

async function getPackageLinks(modulePath) {
  const moduleHtml = await getData(modulePath);
  const root = parse(moduleHtml);

  return root
    .querySelectorAll('#package-summary-table .summary-table .col-first a')
    .map(el => el.getAttribute('href'))
    .map(link => new URL(link, modulePath).toString());
}

async function getClassLinks(packagePath) {
  const packageHtml = await getData(packagePath);
  const root = parse(packageHtml);

  return root
    .querySelectorAll('#class-summary .summary-table .col-first a')
    .map(el => el.getAttribute('href'))
    .map(link => new URL(link, packagePath).toString());
}

async function main() {
  
  // // const t = await genType('https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/lang/String.html')
  // const t = await genType('https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/util/List.html')
  // const types = [ t ];
  // fs.writeFileSync(outFilePath, JSON.stringify(types))

  const basePath = 'https://docs.oracle.com/en/java/javase/17/docs/api/index.html';
  const moduleLinks = await getModuleLinks(basePath);
  console.log(moduleLinks);

  const types = [];
  let successCount = 0;
  let errorCount = 0;

  await Promise.all(moduleLinks.map(async moduleLink => {
    const packageLinks = await getPackageLinks(moduleLink);

    await Promise.all(packageLinks.map(async packageLink => {
      const classLinks = await getClassLinks(packageLink);

      await Promise.all(classLinks.map(async classLink => {
        try {
          const type = await genType(classLink);
          types.push(type);
          successCount++;
        } catch (e) {
          console.error(e);
          errorCount++;
        }
      }));
    }));
  }));

  fs.writeFileSync(outFilePath, JSON.stringify(types));
  console.log(`Successes: ${successCount}, errors: ${errorCount}`)
}

main().catch(console.error)
