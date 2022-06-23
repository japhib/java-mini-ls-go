const { parse, NodeType } = require('node-html-parser');
const axios = require('axios');
const path = require('path');
const fs = require('fs');
const fsPromises = fs.promises;
const util = require('util');

const baseUrl = 'https://docs.oracle.com/en/java/javase/17/docs/api';
const cacheFolder = 'java_sdk_cache';
const outFilePath = path.join(__dirname, '..', 'parse', 'builtins.go');

function toCachedPath(url) {
  if (!url.startsWith(baseUrl)) {
    throw new Error(`url doesn't start with baseUrl: ${url}`);
  }

  const trimmed = url.substring(baseUrl.length);
  return path.join(__dirname, cacheFolder, trimmed);
}

function stripNewlines(text) {
  // the html parser puts a bunch of unnecessary newlines in
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

  const title = root.querySelector('h1.title').textContent;
  if (!title.startsWith('Class ')) {
    throw new Error(`title doesn't start with 'Class '! title: ${title}`);
  }
  data.name = title.substring('Class '.length);

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

  data.fields = parseThreeColumnTable(root, '.field-summary .summary-table > div');
  data.methods = parseThreeColumnTable(root, '.method-summary .summary-table > div.method-summary-table').map(parseArgs);
  data.constructors = parseTwoColumnTable(root, '.constructor-summary .summary-table > div').map(parseArgs);

  console.log(util.inspect(data, {depth: 99}));

  return data;
}

function parseThreeColumnTable(root, selector) {
  const tableChildren = root.querySelectorAll(selector);
  
  const items = [];
  for (let i = 3; i < tableChildren.length - 2; i += 3) {
    const modifierAndType = stripNewlines(tableChildren[i].text);
    // modifierAndType could look like "static final String", in which case we want
    // modifiers to be ["static", "final"] and type to be "string"
    const modifiers = modifierAndType.split(' ');
    const [ type ] = modifiers.splice(-1);

    const name = stripNewlines(tableChildren[i+1].text);
    const description = stripNewlines(tableChildren[i+2].text)

    items.push({
      name, modifiers, type, description
    })
  }

  return items;
}

function parseTwoColumnTable(root, selector) {
  const tableChildren = root.querySelectorAll(selector);
  
  const items = [];
  for (let i = 2; i < tableChildren.length - 1; i += 2) {
    items.push({
      name: stripNewlines(tableChildren[i].text),
      description: stripNewlines(tableChildren[i+1].text),
    })
  }

  return items;
}

function parseArgs(item) {
  const { name } = item;

  const matches = /(.*)\((.*)\)/.exec(name);
  if (!matches) {
    console.log(`Can't parse args, method formatted wrong: ${item}`);
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

async function main() {
  const StringType = await genType('https://docs.oracle.com/en/java/javase/17/docs/api/java.base/java/lang/String.html')
  
  const types = [ StringType ];

  fs.writeFileSync(path.join(__dirname, '..', 'java_stdlib.json'), JSON.stringify(types))
}

main().catch(console.error)
