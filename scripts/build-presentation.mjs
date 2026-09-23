import fs from 'node:fs/promises';
import path from 'node:path';
import {pathToFileURL,fileURLToPath} from 'node:url';
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const modules=process.env.RUNTIME_NODE_MODULES,skill=process.env.PRESENTATIONS_SKILL;
if(!modules||!skill)throw new Error('Set RUNTIME_NODE_MODULES and PRESENTATIONS_SKILL to the installed bundled runtime');
const {Presentation,PresentationFile}=await import(pathToFileURL(path.join(modules,'@oai/artifact-tool/dist/artifact_tool.mjs')).href).catch(async()=>{
 const {createRequire}=await import('node:module');const require=createRequire(path.join(modules,'package.json'));return import(pathToFileURL(require.resolve('@oai/artifact-tool')).href);
});
const {finalizePresentation,applyPresentationChartFont}=await import(pathToFileURL(path.join(skill,'container_tools/artifact_tool_utils.mjs')).href);
const dir=path.join(root,'docs/competition'),tmp=path.join(root,'network/runtime/stage8-presentation');await fs.mkdir(tmp,{recursive:true});
const slides=JSON.parse(await fs.readFile(path.join(dir,'slides.json'),'utf8'));
const p=Presentation.create({slideSize:{width:1280,height:720}}),font='Liberation Sans';
function text(slide,value,x,y,w,h,size=28,bold=false,color='#193B2B'){
 const shape=slide.shapes.add({geometry:'textbox',position:{left:x,top:y,width:w,height:h},fill:'none',line:{fill:'none',width:0}});
 shape.text=value;shape.text.style={typeface:font,fontSize:size,bold,color,autoFit:'none'};return shape;
}
function table(slide,values,y,widths){
 const t=slide.tables.add({rows:values.length,columns:values[0].length,left:65,top:y,width:1150,height:values.length*64,columnWidths:widths,values});
 for(let r=0;r<values.length;r++)for(let c=0;c<values[0].length;c++){
  const cell=t.getCell(r,c);cell.fill=r===0?'#23543B':'#F4F4EB';cell.text.style={typeface:font,fontSize:25,color:r===0?'#FFFFFF':'#193B2B',bold:r===0};
 }
 return t;
}
for(let i=0;i<slides.length;i++){
 const data=slides[i],s=p.slides.add();s.background.fill='#F4F4EB';
 text(s,data.title,65,45,1145,86,i===0?64:46,true);
 text(s,data.lead,65,150,1145,85,30);
 if(data.kind==='architecture'){
  const nodes=[['Aktör / denetçi arayüzü',65,260,480],['Spring Boot API',65,350,480],['Fabric Gateway',65,440,480],['Go chaincode',65,530,480],['Kurum simülatörleri\nSIMULATED',710,335,500],['Backend belge arşivi / günlük',710,440,500],['Ortak ledger / Özel koleksiyonlar',710,545,500]];
  for(const [label,x,y,w]of nodes){const box=s.shapes.add({geometry:'rect',position:{left:x,top:y,width:w,height:65},fill:'#E5EADA',line:{fill:'#23543B',width:1}});box.text=label;box.text.style={typeface:font,fontSize:24,color:'#193B2B',autoFit:'none'};}
  for(const y of [325,415,505])text(s,'↓',285,y,60,26,24);
  text(s,'↔',590,362,70,48,34);text(s,'↔',590,545,70,48,34);
  text(s,'4 peer / 4 MSP · 3 Raft orderer · TLS · tek yerel makine',65,628,1135,36,22);
 }else if(data.kind==='privacy'){
  table(s,[['Alan','Örnek','Yetki'],['Ortak ledger','Parti, sahiplik, kanıt','Kanal üyeleri'],['PDC','Ticari değerler, inceleme','Koleksiyon + rol'],['Zincir dışı','Özgün belge, günlük','Backend denetimi']],255,[340,465,345]);
 }else if(data.kind==='price'){
  const chart=s.charts.add('bar',{position:{left:65,top:252,width:610,height:325},categories:['Normal','Şüpheli'],series:[{name:'Artış (%)',values:[40,85],fill:'#23543B'},{name:'Pilot eşiği (%)',values:[50,50],fill:'#AC672D'}],barOptions:{direction:'column',grouping:'clustered'},hasLegend:true,legend:{position:'bottom',textStyle:{fontSize:20}},xAxis:{textStyle:{fontSize:22}},dataLabels:{showValue:true,position:'outEnd',textStyle:{fontSize:22}}});applyPresentationChartFont(chart,{fontFamily:font});
  text(s,'20 → 28 TL/kg\n%40 artış\nEşik aşımı yok',745,265,470,135,30,true);
  text(s,'20 → 37 TL/kg\n%85 artış\nİnceleme sinyali',745,445,470,135,30,true,'#804213');
 }else if(data.kind==='consumer'){
  for(let j=0;j<data.body.length;j++)text(s,data.body[j],65,265+j*110,560,100,30);
  text(s,'LOT-DEMONORMAL01',700,265,510,50,28,true);
  text(s,'Domates / Antalya\n100 kg / Hasat: 15.09.2026',700,335,510,100,28);
  text(s,'Üretici kaydı\nTaşıyıcı kabulü\nPerakendeci kabulü',700,445,510,135,28,true);
 }else if(data.kind==='measurement'){
  table(s,[['Aşama 7 ölçümü','n','Medyan','p95'],['Gateway → VALID','30','2.022 ms','2.043 ms'],['Ortak sorgu','30','3,0 ms','3,9 ms'],['Tam rota + görünüm','6','16,52 sn','16,59 sn']],260,[505,95,275,275]);
  text(s,'113 farklı test ve iki temiz tekrar PASS. Eşzamanlılık 1.',65,554,1135,45,26,true);
  text(s,'Kapasite/TPS iddiası yoktur. Ölçüm tanımları ve ham kanıt raporda.',65,607,1135,40,22);
 }else{
  const spacing=data.body.length===4?83:105;
  for(let j=0;j<data.body.length;j++)text(s,data.body[j],65,265+j*spacing,1125,90,32,i===0&&j===0);
 }
 if(data.foot)text(s,data.foot,65,630,1145,50,21,false,'#506352');
 text(s,String(i+1).padStart(2,'0'),1190,684,50,25,18);
 s.speakerNotes.textFrame.setText(data.notes);
 const preview=await p.export({slide:s,format:'png',scale:1});await fs.writeFile(path.join(tmp,`slide-${i+1}.png`),new Uint8Array(await preview.arrayBuffer()));
}
const candidate=path.join(tmp,'candidate.pptx');await(await PresentationFile.exportPptx(p)).save(candidate);
await finalizePresentation({workspaceDir:root,candidatePath:candidate,finalPath:path.join(dir,'AgroChain.pptx'),pythonExecutable:process.env.RUNTIME_PYTHON,integrityValidatorPath:path.join(skill,'container_tools/inspect_presentation_package_integrity.py'),layoutValidatorPath:path.join(skill,'container_tools/inspect_presentation_layout_geometry.py'),layoutArgs:['--expected-slide-size-emu','12192000,6858000','--validate-heading-fit','--require-native-table-slide','5','--require-native-table-slide','10'],explicitTotalSlideCount:12,requiredNativeTableOwnerSlides:[5,10],requiredNativeChartOwnerSlides:[7],materializeLiteralChartWorkbooks:true,fontPolicy:{basis:'design',families:[font]},verifyArtifactToolImport:true,receiptPath:path.join(tmp,'validation.json')});
console.log('PASS exported and validated 12-slide presentation');
