## กฏ
- ห้าม git hub copilot มาแก้ไขใน folder spec/ ใดๆ ทั้งสิ้น
- ห้ามสร้าง file .md เพื่ออธิบายขั้นตอนการทดสอบ หรือ API ต่างๆ
- ใช้ภาษาไทยในการอธิบายทั้งหมด
- ใช้ภาษาไทยในการเขียนทั้งหมด
- เขียนให้กระชับ สั้น และได้ใจความ
- Function ที่ไม่ได้ใช้ ให้ลบออกเสมอ ลบ code ที่ไม่จำเป็นออก ทั้งหมด
- ไม่ต้องสร้าง File .md เพื่ออธิบายขั้นตอนการทดสอบ หรือ API ต่างๆ
- ในระบบพยายามเพิ่ม Log เพื่อช่วยในการ Debug
- ไม่ต้องทดสอบ API 
- .go พยายามแยก file ให้เหมาะสม กับหน้าที่ของมัน
- .ps1 run บน Windows PowerShell ไม่ผ่าน ใช้วิธีอื่นแทน

## ระบบ
- พัฒนาโดยภาษา Go
- ใช้ Docker & Docker Compose สำหรับรัน Ollama service
- ระบบนี้ Run บน Docker container เมื่อแก้เสร็จให้ build image ใหม่เสมอ Deploy ขึ้น Desktop Docker เลย
- ใช้ PostgreSQL remote server สำหรับเก็บ vector data
- ใช้ Ollama สำหรับสร้าง embeddings ด้วยโมเดล bge-m3:latest
- ใช้ PostgreSQL พร้อม pgvector extension สำหรับจัดเก็บและค้นหา vector data

## ฟังก์ชัน


### Build Document Vectors
- สำหรับสร้าง vector database ใหม่ผ่าน API
- ให้ไปดึง file จาก `doc/` แล้วสร้าง embeddings ใหม่ ตาม filename และ shopid ที่ส่งมา
- ให้ลบข้อมูลเก่าออกก่อน แล้วเพิ่มข้อมูลใหม่เข้าไป key ได้แก่ shopid, filename เพื่อต้องการแยกกิจการของลูกค้าแต่ละราย และ filename เพื่อระบุเอกสาร
- ให้ประมวลผลพร้อมกัน 100 Threads เพื่อเพิ่มความเร็ว


- **Method**: POST
- **URL**: `http://localhost:8080/build`
- **Headers**: `Content-Type: application/json`
- **Body**:
```json
{
    "shopid": "shop001",
    "filename": "doc01.md",
}
```

### Search ค้นหาได้ทั้งภาษาไทยและอังกฤษ
- รองรับการค้นหาด้วยภาษาไทยและอังกฤษ
- วิธีการค้นหา ให้เอา query ที่รับมา ไปสร้าง embeddings ด้วย Ollama
- จากนั้นนำ vector ที่ได้ ไปค้นหา similarity กับข้อมูลในตาราง documents เพื่อหาข้อมูลที่ใกล้เคียงที่สุด เอาเฉพาะข้อมูลที่ตรงกับ shopid ที่ส่งมา และ similarity สูงกว่า 0.5
- คืนค่าผลลัพธ์เป็น JSON format พร้อมรายละเอียด content, filename, shopid, chunk number, similarity score
- เอาคำที่ตัดได้ ไปหาตามรอบของ array แล้วเอามาประกอบกันเป็นผล ก่อน return

- **Method**: POST  
- **URL**: `http://localhost:8080/search`
- **Headers**: `Content-Type: application/json`
- **Body**:
```json
{
    "query": "วันหยุดมีกี่วัน",
    "shopid": "shop001",
    "limit": 3
}
```
