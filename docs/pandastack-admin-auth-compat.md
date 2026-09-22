# Admin Authorization header compatibility

This opt-in-by-code compatibility layer is intended for hosting proxies that drop the standard Authorization header. It was prepared against Tazimi1012/ds2api commit 1470eef7a20a9b1eb22907445118f0b311d08ab5.

## Behavior

- The WebUI fetch wrapper adds `X-Ds2api-Admin-Authorization` alongside the original Authorization header, only for authenticated, same-origin `/admin` or `/admin/*` requests.
- No credentials are stored by the wrapper. Non-admin and external requests remain unchanged. Authenticated admin redirects are rejected to prevent forwarding a custom credential header to a redirect target.
- Middleware is installed only in the `/admin` router. It reads the fallback only when standard Authorization is empty. An existing standard header always takes precedence, even if invalid.
- Original password/JWT verification is unchanged. Invalid or missing credentials remain rejected. There are no anonymous-admin exceptions, URL tokens, or hard-coded production credentials.
- No behavior changes to `/v1` or other public API protocols. This is not a fix for external OpenAI-compatible clients whose Authorization header is also stripped.

## Deployment

Deploy both backend and frontend changes together and rebuild the WebUI. The existing PandaStack build/start commands and Secrets stay in use. Hard refresh the admin page after deployment. Verify login, a page refresh, settings access, and rejection of invalid/missing credentials.

This fallback must survive the platform proxy to work. If the proxy also removes it, or replaces standard Authorization with a non-empty value, do not disable authentication or keep adding unverified fallbacks; investigate the proxy path.

## Tests

Run `node --test webui/src/utils/adminAuthHeaderCompat.test.js` and `go test ./internal/server ./internal/auth ./internal/httpapi/admin/auth`, plus the repository quality gates.

## Rollback

Revert the patch commit and redeploy. The old standard-Authorization-only flow returns; the original hosting compatibility problem may also return. Do not remove only half of the paired frontend/backend implementation.
