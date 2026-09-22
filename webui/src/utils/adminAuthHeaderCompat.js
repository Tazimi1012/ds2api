// Preserve standard Authorization and provide a fallback for proxies that strip it.
// Only authenticated same-origin /admin requests get the fallback; no token storage.
export const ADMIN_AUTH_HEADER = 'X-Ds2api-Admin-Authorization'

export function createAdminAuthFetch(nativeFetch, baseURL) {
    const base = new URL(baseURL)
    return function adminAuthFetch(input, init) {
        const request = typeof Request !== 'undefined' && input instanceof Request ? input : null
        const url = new URL(request ? request.url : String(input), base)
        const isAdmin = url.pathname === '/admin' || url.pathname.startsWith('/admin/')
        if (url.origin !== base.origin || !isAdmin) return nativeFetch(input, init)

        const headers = new Headers(init?.headers ?? request?.headers)
        const authorization = headers.get('Authorization')
        if (!authorization || !/^Bearer\s+\S+/i.test(authorization.trim())) {
            return nativeFetch(input, init)
        }
        headers.set(ADMIN_AUTH_HEADER, authorization)
        // Custom headers are not always stripped on cross-origin redirects.
        // Fail closed on redirects instead of risking credential forwarding.
        return nativeFetch(input, { ...init, headers, redirect: 'error' })
    }
}

const installed = Symbol.for('ds2api.adminAuthHeaderCompat.installed')
export function installAdminAuthHeaderCompat(target = globalThis) {
    if (target[installed]) return
    target.fetch = createAdminAuthFetch(target.fetch.bind(target), target.location.href)
    target[installed] = true
}
