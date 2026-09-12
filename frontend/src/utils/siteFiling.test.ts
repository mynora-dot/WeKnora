import assert from 'node:assert/strict'
import { test } from 'node:test'
import { execFileSync, spawnSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { runInNewContext } from 'node:vm'
import { resolveSiteFiling } from './siteFiling'

const complete = {
  PUBLIC_SECURITY_BEIAN_NUMBER: '京公网安备11000000000000号',
  PUBLIC_SECURITY_BEIAN_URL: 'https://beian.mps.gov.cn/',
  PUBLIC_SECURITY_BEIAN_ICON_URL: '/beian.png',
}
test('empty, ICP only, police only and both configurations', () => {
  assert.equal(resolveSiteFiling().visible, false)
  assert.equal(resolveSiteFiling({ ICP_BEIAN_NUMBER: '  ' }).visible, false)
  assert.deepEqual(resolveSiteFiling({ ICP_BEIAN_NUMBER: ' 京ICP备测试号 ' }), {
    icp: '京ICP备测试号', police: null, visible: true,
  })
  assert.ok(resolveSiteFiling(complete).police)
  assert.equal(resolveSiteFiling({ ...complete, ICP_BEIAN_NUMBER: '测试' }).icp, '测试')
  for (const key of Object.keys(complete)) {
    assert.equal(resolveSiteFiling({ ...complete, [key]: '' }).police, null)
  }
})
test('reject unsafe addresses and accept HTTPS icons', () => {
  for (const url of ['javascript:alert(1)', 'data:text/html,test', '//evil.test', 'http://example.com', 'https://user:pass@example.com', 'https://', 'https://exam\nple.com/']) {
    assert.equal(resolveSiteFiling({ ...complete, PUBLIC_SECURITY_BEIAN_URL: url }).police, null)
  }
  for (const icon of ['//evil.test/icon', '/\\evil.test/icon', 'data:image/png,test', 'javascript:alert(1)']) {
    assert.equal(resolveSiteFiling({ ...complete, PUBLIC_SECURITY_BEIAN_ICON_URL: icon }).police, null)
  }
  assert.ok(resolveSiteFiling({ ...complete, PUBLIC_SECURITY_BEIAN_ICON_URL: 'https://example.com/beian.png' }).police)
})
test('container serializer preserves strings, exposes only filing fields, updates and clears values', () => {
  const special = '中文 " \\ \n\r\t </script>\u2028\u2029 ;globalThis.injected=true;'
  const defaults = readFileSync(new URL('../../public/config.js', import.meta.url), 'utf8')
  const env = { PATH: process.env.PATH, PRIVATE_SECRET: 'must-not-leak', ICP_BEIAN_NUMBER: special }
  const script = new URL('../../filing-config.sh', import.meta.url).pathname
  const output = execFileSync('sh', [script], { env, encoding: 'utf8' })
  const context = { window: {} as { __RUNTIME_CONFIG__?: Record<string, string> } }
  runInNewContext(defaults + output, context)
  assert.equal(context.window.__RUNTIME_CONFIG__?.ICP_BEIAN_NUMBER, special)
  assert.ok(!output.includes('must-not-leak'))
  assert.equal((context as any).injected, undefined)
  const cleared = execFileSync('sh', [script], { env: { PATH: process.env.PATH }, encoding: 'utf8' })
  runInNewContext(cleared, context)
  assert.equal(context.window.__RUNTIME_CONFIG__?.ICP_BEIAN_NUMBER, '')
  const partial = spawnSync('sh', [script], { env: { PATH: process.env.PATH, PUBLIC_SECURITY_BEIAN_NUMBER: '测试' }, encoding: 'utf8' })
  assert.equal(partial.status, 0)
  assert.match(partial.stderr, /incomplete/)
})
