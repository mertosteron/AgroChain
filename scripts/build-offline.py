#!/usr/bin/env python3
"""Build a self-contained, clearly non-live presentation from reviewed local files."""
import base64
import html
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs/competition"


def build():
    slides = json.loads((OUT / "slides.json").read_text())
    assert len(slides) == 12
    sections = []
    for number, slide in enumerate(slides, 1):
        body = "".join("<p>" + html.escape(x) + "</p>" for x in slide.get("body", []))
        if slide.get("kind") == "architecture":
            body = '<div class="flow">Aktör / denetçi arayüzü<br>↓<br>Spring Boot API ↔ İmzalı kurum simülatörleri (SIMULATED)<br>↓<br>Fabric Gateway · kuruluş kimliği<br>↓<br>Go chaincode · deterministik doğrulama<br>↓<br>Ortak ledger ↔ Üç özel veri koleksiyonu</div><p>Backend: belge arşivi ve kalıcı işlem günlüğü. Geçerli olaylardan tüketici görünümü. 4 peer, 3 Raft orderer, TLS.</p>'
        sections.append(f'<section class="slide" id="slide-{number}" tabindex="-1" {"hidden" if number > 1 else ""}><div class="number">AgroChain / {number:02d}</div><h1>{html.escape(slide["title"])}</h1><h2>{html.escape(slide["lead"])}</h2>{body}<p class="foot">{html.escape(slide.get("foot", ""))}</p><details><summary>Konuşmacı notları ve kaynak</summary><p>{html.escape(slide["notes"])}</p></details></section>')
    images = []
    for filename, label in [("comparison.png", "Gerçek yerel testten kayıtlı normal / şüpheli parti karşılaştırması"), ("consumer.png", "Gerçek yerel testten kayıtlı tüketici görünümü")]:
        data = base64.b64encode((OUT / "assets" / filename).read_bytes()).decode()
        images.append(f'<figure><figcaption>{label}</figcaption><img alt="{label}" src="data:image/png;base64,{data}"></figure>')
    document = '''<!doctype html><html lang="tr"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; img-src data:; connect-src 'none'; base-uri 'none'"><title>AgroChain / Çevrimdışı yarışma paketi</title><style>
    :root{font-family:Arial,sans-serif;color:#193b2b;background:#f4f4eb}*{box-sizing:border-box}body{margin:0}.warning{padding:16px 4vw;background:#804213;color:white;font-weight:bold;position:sticky;top:0;z-index:1}.slide{padding:35px 6vw;max-width:1440px;min-height:75vh;margin:auto}h1{font-size:clamp(32px,4vw,60px);line-height:1.1;max-width:1100px}h2{font-size:clamp(22px,2.4vw,34px);font-weight:normal}p,.flow{font-size:clamp(18px,2vw,28px);line-height:1.5}.number,.foot,details p{font-size:18px}.foot{margin-top:30px;color:#506352}nav{padding:20px 6vw;display:flex;gap:16px;align-items:center;flex-wrap:wrap}button,a{font:inherit;min-height:44px;padding:12px;background:#23543b;color:white;border:0;border-radius:3px}a:focus-visible,button:focus-visible,summary:focus-visible{outline:3px solid #804213;outline-offset:3px}button:disabled{opacity:.45}#evidence{padding:35px 6vw;max-width:1440px;margin:auto}figure{margin:30px 0}img{max-width:100%;height:auto}figure:last-child img{max-height:1000px;width:auto}figcaption{font-size:22px;margin-bottom:15px}details{margin-top:24px}summary{cursor:pointer;min-height:44px}details p{line-height:1.5}@media print{.slide[hidden]{display:block}.warning{position:static}.slide{break-after:page}nav{display:none}details{display:none}}
    </style><div class="warning">ÇEVRİMDIŞI KAYIT • Canlı sorgu veya yeni zincir işlemi yapılmaz • Kurum verileri SIMULATED</div><main>SLIDES</main><nav aria-label="Slayt gezintisi"><button id="prev">← Önceki</button><span id="position" aria-live="polite">1 / 12</span><button id="next">Sonraki →</button><a href="#evidence">Kayıtlı ekranlar</a></nav><section id="evidence"><h1>Kayıtlı sentetik demo</h1><p>Bu ekranlar gerçek yerel Fabric testi sırasında sabit demo partilerinden alındı. Canlı ağın güncel durumunu göstermez. Gizli anahtar veya token içermez.</p>IMAGES<a href="#slide-1" id="back">Sunuma dön</a></section><script>
    const slides=[...document.querySelectorAll('.slide')];let current=0;
    function show(n){current=Math.max(0,Math.min(slides.length-1,n));slides.forEach((s,i)=>s.hidden=i!==current);document.getElementById('position').textContent=(current+1)+' / '+slides.length;document.getElementById('prev').disabled=current===0;document.getElementById('next').disabled=current===slides.length-1;slides[current].focus({preventScroll:true});window.scrollTo(0,0);}
    document.getElementById('prev').onclick=()=>show(current-1);document.getElementById('next').onclick=()=>show(current+1);document.getElementById('back').onclick=()=>show(0);document.addEventListener('keydown',e=>{if(e.target.closest('summary,details'))return;if(e.key==='ArrowRight'){e.preventDefault();show(current+1);}if(e.key==='ArrowLeft'){e.preventDefault();show(current-1);}if(e.key==='Home')show(0);if(e.key==='End')show(slides.length-1);});show(0);
    </script></html>'''.replace("SLIDES", "".join(sections)).replace("IMAGES", "".join(images))
    (OUT / "offline.html").write_text(document)
    print("PASS self-contained 12-slide offline HTML generated")


if __name__ == "__main__":
    build()
