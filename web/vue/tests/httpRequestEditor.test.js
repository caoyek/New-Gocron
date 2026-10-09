const assert = require('assert')
const fs = require('fs')
const path = require('path')

const source = fs.readFileSync(path.join(__dirname, '../src/pages/task/httpRequestEditor.vue'), 'utf8')
const script = source.match(/<script>([\s\S]*?)<\/script>/)[1]
const component = new Function(script.replace('export default', 'return'))()

function editor (overrides) {
  const vm = Object.assign(component.data(), {
    form: Object.assign({command: '', http_method: 2, http_headers: '', http_body_mode: '', http_body: ''}, overrides),
    $message: {error: () => {}}
  })
  Object.keys(component.methods).forEach(name => { vm[name] = component.methods[name].bind(vm) })
  vm.readParams()
  vm.readHeaders()
  return vm
}

const params = editor({command: 'https://example.invalid/api?a=hello%20world&a=2#section'})
assert.strictEqual(params.form.command, 'https://example.invalid/api?a=hello%20world&a=2#section')
assert.deepStrictEqual(params.params.map(item => [item.name, item.value]), [['a', 'hello world'], ['a', '2']])
params.params[0].value = '中文 & space'
params.writeParams()
assert.strictEqual(params.form.command, 'https://example.invalid/api?a=%E4%B8%AD%E6%96%87+%26+space&a=2#section')
const rowId = params.params[0].id
params.readParams()
assert.strictEqual(params.params[0].id, rowId, 'typing must not replace focused rows')
params.removeParam(1)
assert.strictEqual(new URL(params.form.command).searchParams.getAll('a').length, 1)

const headers = editor({http_headers: '{"Authorization":"Bearer example"}'})
assert.strictEqual(headers.headers[0].value, 'Bearer example')
headers.addRow(headers.headers)
Object.assign(headers.headers[1], {name: 'authorization', value: 'duplicate'})
assert.strictEqual(headers.validate(), false)
headers.headers.pop()
assert.strictEqual(headers.validate(), true)
assert.deepStrictEqual(JSON.parse(headers.form.http_headers), {Authorization: 'Bearer example'})

const broken = editor({http_headers: '{broken'})
assert.strictEqual(broken.validate(), false)
assert.strictEqual(broken.form.http_headers, '{broken', 'invalid stored settings must not be silently erased')
const json = editor({http_body_mode: 'json', http_body: '{"ok":true}'})
assert.strictEqual(json.validate(), true)
json.formatJSON()
assert.strictEqual(json.form.http_body, '{\n  "ok": true\n}')
json.form.http_body = '{broken'
assert.strictEqual(json.validate(), false)
const empty = editor({http_body_mode: 'json', http_body: '   '})
assert.strictEqual(empty.validate(), true)
empty.formatJSON()
assert.strictEqual(empty.form.http_body, '   ')
const legacy = editor({http_body_mode: '', command: 'https://example.invalid/api?a=1'})
legacy.legacyForm = true
assert.strictEqual(legacy.validate(), true)
assert.strictEqual(legacy.form.http_body_mode, '')
console.log('HTTP editor checks passed: URL roundtrip, duplicate parameters, stable rows, Headers, invalid data preservation, JSON validation/formatting')
