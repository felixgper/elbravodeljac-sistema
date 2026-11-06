import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import {validateIsolatedFixture} from './isolated-fixture.mjs';

const token='a'.repeat(64);let requests=0;
const fakeRequest=async()=>{requests++;throw Error('Unexpected network');};
for(const url of ['http://localhost:8080','http://127.0.0.1:8080','https://example.com','http://127.0.0.1:9999/private','http://user:secret@127.0.0.1:9999','http://127.0.0.1:9999/?next=8080']) {
  await assert.rejects(validateIsolatedFixture({url,isolationToken:token},fakeRequest),/QA_ISOLATION_REQUIRED/);
}
await assert.rejects(validateIsolatedFixture({url:'http://127.0.0.1:9999'},fakeRequest),/QA_ISOLATION_REQUIRED/);
assert.equal(requests,0,'Unsafe targets must fail before any HTTP request');
await assert.rejects(validateIsolatedFixture({url:'http://127.0.0.1:9999',isolationToken:token},async()=>({status:200,headers:new Headers()})),/QA_ISOLATION_REQUIRED/);
// Unit test only: the real HTTP capability is exercised by the Go/Chrome suites.
const f=await validateIsolatedFixture({url:'http://127.0.0.1:9999',isolationToken:token},async()=>({status:200,headers:new Headers({'X-Bravo-Test-Run':token})}));
assert(f.testRunId&&!f.testRunId.includes(token));
// Sentinel module must never be loaded: rejection precedes Playwright/private files.
for(const script of ['attention-browser.mjs','brands-browser.mjs','diagnosis-browser.mjs','quotes-browser.mjs','review-browser.mjs','live-docker.mjs','read-persisted-docker.mjs','audit-browser.mjs','frontend-browser.mjs','carbon-design-browser.mjs','sheet-functional-browser.mjs','taller-browser.mjs','clean-copy-browser.mjs']) {
  const result=spawnSync(process.execPath,['checks/'+script],{input:JSON.stringify({url:'http://localhost:8080',isolationToken:token}),encoding:'utf8',timeout:5000,env:{...process.env,PLAYWRIGHT_MODULE:'must-not-be-loaded'}});
  assert.notEqual(result.status,0,script);
  assert.match(result.stderr,/QA_ISOLATION_REQUIRED/,script);
  assert(!result.stderr.includes('MODULE_NOT_FOUND'),script+' loaded Playwright before rejecting the target');
}
console.log('Aislamiento: destinos peligrosos rechazados sin HTTP, servidor de prueba verificado y 13 entradas bloqueadas antes de Playwright/credenciales. PASS (unitario, sin bases de datos).');
