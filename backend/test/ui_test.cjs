/* Real browser -> HTTP -> Fabric smoke; no mocked API responses or embedded keys. */
const {chromium}=require('playwright');
const fs=require('node:fs'),path=require('node:path'),{spawn}=require('node:child_process'),assert=require('node:assert/strict');
const root=path.resolve(__dirname,'../..'),runtime=path.join(root,'network/runtime/backend-ui-acceptance');
const tokens=JSON.parse(fs.readFileSync(path.join(root,'network/runtime/backend/tokens.json'),'utf8'));
const fixtures=JSON.parse(fs.readFileSync(path.join(root,'network/runtime/backend-acceptance/stage6-summary.json'),'utf8'));
fs.mkdirSync(runtime,{recursive:true,mode:0o700});
const log=fs.openSync(path.join(runtime,'server.log'),'a',0o600),port=18081,origin=`http://127.0.0.1:${port}`;
const server=spawn(path.join(root,'network/tools/jdk-21.0.12.1+1/bin/java'),['-jar',path.join(root,'backend/target/agrochain-backend-0.3.0.jar')],{cwd:root,env:{...process.env,AGROCHAIN_ROOT:root,AGROCHAIN_BACKEND_DATA:runtime,AGROCHAIN_TOKEN_FILE:path.join(root,'network/runtime/backend/tokens.json'),AGROCHAIN_API_PORT:String(port)},stdio:['ignore',log,log]});
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
async function eventually(fn){let last;for(let i=0;i<100;i++){try{return await fn();}catch(e){last=e;await sleep(300);}}throw last;}
let browser;
(async()=>{
 await eventually(async()=>assert.equal((await fetch(origin+'/api/v1/health')).status,200));
 browser=await chromium.launch({headless:true,...(process.env.CHROMIUM_PATH?{executablePath:process.env.CHROMIUM_PATH}:{})});
 const page=await browser.newPage({viewport:{width:1280,height:1000}}),errors=[];
 page.on('pageerror',e=>errors.push(e.message));page.setDefaultTimeout(20000);
 async function login(who){await page.goto(origin);await page.locator('#token').fill(tokens[who]);await page.getByRole('button',{name:'Çalışma alanını aç →'}).click();await page.locator('#workspace').waitFor({state:'visible'});await page.locator('#message').filter({hasText:'Kurumsal oturum açıldı.'}).waitFor();}
 async function choose(batch){await page.locator('#search').fill(batch);await eventually(async()=>{await page.getByRole('button',{name:'Kayıtları yenile'}).click();const b=page.locator('#batches button');assert.equal(await b.count(),1);});await page.locator('#batches button').click();await page.locator('#detail h2').filter({hasText:'Domates / Antalya'}).waitFor();}
 async function commit(label){const response=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().includes('/api/v1/'));await page.getByRole('button',{name:label,exact:true}).click();const r=await response;const body=await r.json();assert.ok(r.ok(),JSON.stringify(body));assert.equal(body.status,'COMMITTED');await page.locator('#message').filter({hasText:'Kayıtlar güncel.'}).waitFor();return body;}
 await login('regulator:reviewer');
 await eventually(async()=>{await page.getByRole('button',{name:'Kayıtları yenile'}).click();assert.ok(await page.locator(`#left option[value="${fixtures[0].batch}"]`).count());});
 await page.selectOption('#left',fixtures[0].batch);await page.selectOption('#right',fixtures[1].batch);await page.getByRole('button',{name:'Karşılaştır',exact:true}).click();
 await page.locator('#compare-results').getByText('Artış 40% · Pilot eşiği 50%',{exact:true}).waitFor();await page.locator('#compare-results').getByText('Artış 85% · Pilot eşiği 50%',{exact:true}).waitFor();
 await page.locator('#comparison').screenshot({path:path.join(runtime,'comparison.png')});
 console.log('PASS inspector compares real 40% and 85% results and histories');
 await login('producer:producer');const created=await commit('Partiyi oluştur'),batch=created.receipt.batchId;await choose(batch);await commit('Taşıyıcıya teslim teklif et');
 await login('logistics:carrier');await choose(batch);await commit('Teslim almayı onayla');await commit('Nakliye belgesini kaydet');await commit('Mağazaya teslim teklif et');
 assert.equal(await page.getByText('Alış / kg',{exact:true}).count(),0);
 await login('retailer:retailer');await choose(batch);await commit('Mağazada teslim al');await page.getByLabel('Raf fiyatı (TL/kg)').fill('37,00');await commit('Raf fiyatını bildir');
 await login('regulator:reviewer');await choose(batch);await eventually(async()=>{await page.getByRole('button',{name:'Kayıtları yenile'}).click();await page.getByRole('button',{name:'İncelemeyi başlat',exact:true}).waitFor({timeout:1500});});await commit('İncelemeyi başlat');
 await page.getByLabel('Gerekçe (kişisel veri yazmayın)').fill('Fiyat farkı kalite belgeleriyle incelendi. Pilot değerlendirmesidir.');await commit('İncelemeyi sonuçlandır');await page.getByRole('heading',{name:'Sonuçlandırıldı',exact:true}).waitFor();
 await page.locator('#detail').screenshot({path:path.join(runtime,'review.png')});
 console.log('PASS browser completes seven actor transactions and two private review actions');
 for(const width of [1280,768,375]){
  await page.setViewportSize({width,height:900});
  await page.goto(origin+'/consumer.html?lot='+fixtures[0].lot);await page.locator('#trace h2').waitFor();
  assert.equal(await page.locator('#trace .timeline li').count(),3);
  const text=await page.locator('#trace').innerText();assert.ok(!/kuruş|fiyat|REVIEW_REQUIRED|NO_SIGNAL|SaltHex|explanation/i.test(text));
  assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth>window.innerWidth),false);
  await page.screenshot({path:path.join(runtime,`consumer-${width}.png`),fullPage:true});
 }
 await page.goto(origin+'/consumer.html?lot=LOT-MISSING1');await page.getByText('Kayıt henüz bulunamadı.',{exact:false}).waitFor();
 await page.setViewportSize({width:375,height:900});await login('regulator:reviewer');await choose(fixtures[1].batch);assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth),false);
 await page.locator('#detail .qr').waitFor();assert.equal(await page.locator('#detail .qr').count(),1);assert.equal(await page.getByRole('heading',{name:'Sonuçlandırıldı',exact:true}).count(),1);
 await page.screenshot({path:path.join(runtime,'review-mobile.png'),fullPage:true});
 await page.getByRole('button',{name:'Çıkış',exact:true}).click();assert.equal(await page.locator('#token').inputValue(),'');assert.equal(await page.evaluate(()=>localStorage.length+sessionStorage.length),0);
 assert.deepEqual(errors,[]);
 console.log('PASS consumer privacy, 375/768/1280 layouts, missing lot, logout and zero browser exceptions');
})().catch(e=>{console.error(e);process.exitCode=1;}).finally(async()=>{if(browser)await browser.close();server.kill('SIGTERM');fs.closeSync(log);});
