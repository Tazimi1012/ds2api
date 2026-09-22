import test from 'node:test'
import assert from 'node:assert/strict'
import { ADMIN_AUTH_HEADER, createAdminAuthFetch, installAdminAuthHeaderCompat } from './adminAuthHeaderCompat.js'

function mock() {
    const calls = []
    const fetch = createAdminAuthFetch((...args) => { calls.push(args); return Promise.resolve('ok') }, 'https://example.test/admin')
    return { fetch, calls }
}
for (const path of ['/admin/verify', '/admin/settings', '/admin/config', '/admin']) {
    test(`adds fallback only for authorized admin request: ${path}`, async () => {
        const { fetch, calls } = mock()
        const init = { headers: { Authorization: 'Bearer test-token' }, method: 'POST', body: '{}' }
        await fetch(path, init)
        assert.equal(calls[0][1].headers.get(ADMIN_AUTH_HEADER), 'Bearer test-token')
        assert.equal(calls[0][1].headers.get('Authorization'), 'Bearer test-token')
        assert.equal(calls[0][1].redirect, 'error')
        assert.equal(calls[0][1].body, '{}')
        assert.equal(init.headers[ADMIN_AUTH_HEADER], undefined)
    })
}
for (const path of ['/v1/chat/completions', '/administrator', 'https://other.test/admin/verify', '//other.test/admin']) {
    test(`leaves non-admin/external request unchanged: ${path}`, async () => {
        const { fetch, calls } = mock()
        const init = { headers: { Authorization: 'Bearer test-token' } }
        await fetch(path, init)
        assert.equal(calls[0][1], init)
        assert.equal(init.headers[ADMIN_AUTH_HEADER], undefined)
    })
}
test('login request without authorization is unchanged', async () => {
    const { fetch, calls } = mock()
    const init = { method: 'POST', body: '{}', headers: { 'Content-Type': 'application/json' } }
    await fetch('/admin/login', init)
    assert.equal(calls[0][1], init)
})
test('Request inputs preserve signal and do not consume body', async () => {
    const { fetch, calls } = mock()
    const req = new Request('https://example.test/admin/config', { method: 'POST', body: '{}', headers: { authorization: 'Bearer req-token' } })
    await fetch(req)
    assert.equal(calls[0][0], req)
    assert.equal(req.bodyUsed, false)
    assert.equal(calls[0][1].headers.get(ADMIN_AUTH_HEADER), 'Bearer req-token')
})
test('init.headers overrides Request headers', async () => {
    const { fetch, calls } = mock()
    const req = new Request('https://example.test/admin/config', { headers: { authorization: 'Bearer old' } })
    await fetch(req, { headers: { authorization: 'Bearer new' } })
    assert.equal(calls[0][1].headers.get(ADMIN_AUTH_HEADER), 'Bearer new')
})
test('unrecognized scheme is not copied', async () => {
    const { fetch, calls } = mock()
    const init = { headers: { Authorization: 'Basic dummy' } }
    await fetch('/admin/verify', init)
    assert.equal(calls[0][1], init)
})
test('installer is idempotent', async () => {
    const target = { fetch: async () => 'ok', location: { href: 'https://example.test/admin' } }
    installAdminAuthHeaderCompat(target)
    const first = target.fetch
    installAdminAuthHeaderCompat(target)
    assert.equal(target.fetch, first)
    assert.equal(await target.fetch('/healthz'), 'ok')
})
