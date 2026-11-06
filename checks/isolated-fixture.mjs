import {createHash} from 'node:crypto';

const blocked = () => new Error('QA_ISOLATION_REQUIRED: ejecutar mediante un harness Go y una base temporal nueva; localhost:8080 y credenciales de desarrollo están prohibidos.');

export async function validateIsolatedFixture(f, request = fetch) {
  let url;
  try { url = new URL(f?.url); } catch { throw blocked(); }
  if (url.protocol !== 'http:' || url.hostname !== '127.0.0.1' || !url.port || url.port === '8080' || url.username || url.password || url.pathname !== '/' || url.search || url.hash || !/^[a-f0-9]{64}$/.test(f?.isolationToken ?? '')) throw blocked();
  let response;
  try { response = await request(new URL('/healthz', url), {redirect:'error', signal:AbortSignal.timeout(5000)}); } catch { throw blocked(); }
  if (response.status !== 200 || response.headers.get('X-Bravo-Test-Run') !== f.isolationToken) throw blocked();
  return {...f, url:url.origin, testRunId:createHash('sha256').update(f.isolationToken).digest('hex').slice(0,16)};
}

export async function readIsolatedFixture() {
  if (process.stdin.isTTY) throw blocked();
  let raw='';
  for await (const chunk of process.stdin) {
    raw += chunk;
    if (raw.length > 16384) throw blocked();
  }
  let f;
  try { f=JSON.parse(raw); } catch { throw blocked(); }
  return validateIsolatedFixture(f);
}
