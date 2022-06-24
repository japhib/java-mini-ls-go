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

function stripNewlines(text) {
  // the html parser puts a bunch of unnecessary newlines in

  // FYI `replaceAll` only exists in Node v16+
  text = text.replaceAll('\n', '');

  // '\xa0' is non-breaking space
  text = text.replaceAll('\xa0', ' ');

  return text;
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

  console.log('Getting type data for class ' + data.name);

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

  data.fields = parseTable('field', root, '.field-summary .summary-table');
  data.constructors = parseTable('constructor', root, '.constructor-summary .summary-table').map(parseArgs);

  data.methods = parseTable('method', root, '.method-summary .summary-table').map(parseArgs);
  if (!data.methods || data.methods.length === 0) {
    data.methods = parseTable('method', root, '#method-summary-table .summary-table').map(parseArgs);
  }

  // console.log(util.inspect(data, {depth: 99}));

  return data;
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
      let content = stripNewlines(tableChildren[i + j].textContent);

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
  const { name } = item;

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
  // const t = await genType('https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/lang/String.html')
  // const types = [ t ];
  // fs.writeFileSync(outFilePath, JSON.stringify(types))

  const baseModulePath = 'https://docs.oracle.com/en/java/javase/17/docs/api/java.base/module-summary.html';
  const packageLinks = await getPackageLinks(baseModulePath);
  console.log(packageLinks);

  let successCount = 0;
  let errorCount = 0;
  
  const types = [];
  for (const packageLink of packageLinks) {
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
  }

  fs.writeFileSync(outFilePath, JSON.stringify(types));

  console.log(`Successes: ${successCount}, errors: ${errorCount}`)
}

main().catch(console.error)
