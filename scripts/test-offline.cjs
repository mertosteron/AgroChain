const {chromium}=require('playwright');
const path=require('node:path'),{pathToFileURL}=require('node:url'),fs=require('node:fs'),assert=require('node:assert/strict');
const root=path.resolve(__dirname,'..');
(async()=>{
 const browser=await chromium.launch({headless:true,...(process.env.CHROMIUM_PATH?{executablePath:process.env.CHROMIUM_PATH}:{})});
 try{
  const context=await browser.newContext({offline:true,viewport:{width:1280,height:900}}),page=await context.newPage(),remote=[],errors=[];
  page.on('request',r=>{if(/^https?:/.test(r.url()))remote.push(r.url());});page.on('pageerror',e=>errors.push(e.message));
  await page.goto(pathToFileURL(path.join(root,'docs/competition/offline.html')).href);
  assert.equal(await page.locator('.slide').count(),12);
  for(let i=1;i<=12;i++){
   assert.equal(await page.locator('.slide:visible').count(),1);
   assert.equal(await page.locator('#position').innerText(),`${i} / 12`);
   assert.ok((await page.locator('.warning').innerText()).includes('ÇEVRİMDIŞI'));
   assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth),false);
   if(i<12)await page.locator('#next').click();
  }
  assert.equal(await page.locator('#next').isDisabled(),true);
  await page.keyboard.press('Home');assert.equal(await page.locator('#position').innerText(),'1 / 12');
  await page.getByRole('link',{name:'Kayıtlı ekranlar',exact:true}).click();
  assert.equal(await page.locator('#evidence img').count(),2);
  assert.ok(await page.locator('#evidence img').evaluateAll(imgs=>imgs.every(i=>i.complete&&i.naturalWidth>0)));
  for(const width of [375,768,1280]){await page.setViewportSize({width,height:900});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth),false);}
  const dir=path.join(root,'network/runtime/stage8-offline');fs.mkdirSync(dir,{recursive:true});
  await page.locator('#back').click();await page.screenshot({path:path.join(dir,'offline.png')});
  assert.deepEqual(remote,[]);assert.deepEqual(errors,[]);
  console.log('PASS offline file: 12 slides, keyboard navigation, embedded screenshots, 375/768/1280, zero HTTP requests and JS errors');
 }finally{await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1;});
