/* Only captures the two explicit synthetic Stage 8 fixtures. Never exports tokens. */
const {chromium}=require('playwright');
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const root=path.resolve(__dirname,'..'),out=path.join(root,'docs/competition/assets');
(async()=>{
 fs.mkdirSync(out,{recursive:true});
 const tokens=JSON.parse(fs.readFileSync(path.join(root,'network/runtime/backend/tokens.json')));
 const browser=await chromium.launch({headless:true,...(process.env.CHROMIUM_PATH?{executablePath:process.env.CHROMIUM_PATH}:{})});
 try{
  const page=await browser.newPage({viewport:{width:1440,height:1200}}),errors=[];
  page.on('pageerror',e=>errors.push(e.message));
  await page.goto('http://127.0.0.1:8080');
  await page.locator('#token').fill(tokens['regulator:reviewer']);
  await page.getByRole('button',{name:'Çalışma alanını aç →'}).click();
  await page.locator('#message').filter({hasText:'Kurumsal oturum açıldı.'}).waitFor();
  await page.selectOption('#left','BAT-DEMONORMAL01');await page.selectOption('#right','BAT-DEMOSUSPICIOUS01');
  await page.getByRole('button',{name:'Karşılaştır',exact:true}).click();
  await page.getByText('Artış 40% · Pilot eşiği 50%',{exact:true}).waitFor();
  await page.getByText('Artış 85% · Pilot eşiği 50%',{exact:true}).waitFor();
  await page.waitForFunction(()=>!document.getElementById('message').textContent);
  await page.locator('#comparison').scrollIntoViewIfNeeded();
  const rect=await page.locator('#comparison').boundingBox();
  await page.screenshot({path:path.join(out,'comparison.png'),clip:{x:rect.x,y:rect.y,width:rect.width,height:Math.min(850,rect.height)}});
  await page.setViewportSize({width:600,height:1500});
  await page.goto('http://127.0.0.1:8080/consumer.html?lot=LOT-DEMONORMAL01');
  await page.locator('#trace h2').waitFor();
  assert.equal(await page.locator('#trace .timeline li').count(),3);
  assert.ok(!/Kurus|classification|saltHex|explanation/.test(await page.locator('#trace').innerText()));
  await page.screenshot({path:path.join(out,'consumer.png'),fullPage:true});
  assert.deepEqual(errors,[]);
  console.log('PASS captured fixed synthetic comparison and public consumer view');
 } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1;});
