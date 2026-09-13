# Auth Project

ตัวอย่างระบบยืนยันตัวตนด้วย Go ประกอบด้วย User Service, PostgreSQL, Kong API Gateway และ ngrok สำหรับเปิด Kong ให้โลกภายนอกเรียกใช้งาน

## ภาพรวมระบบ

```text
Client ภายนอก
      |
      v
ngrok (Public URL)
      |
      v
Kong :8000
  |-- /auth/*   -> user-service:8080
  |-- /chaiyos  -> Chaiyos ผ่าน ngrok URL ใน kong.yml
      |
PostgreSQL :5432
```

Docker Compose จะรันบริการต่อไปนี้พร้อมกัน:

- `postgres-db`: ฐานข้อมูล PostgreSQL
- `user-service`: สมัครสมาชิกและเข้าสู่ระบบ พร้อมออก JWT
- `kong`: API Gateway และตรวจสอบ JWT สำหรับ `/chaiyos`
- `ngrok`: เปิด Kong `kong:8000` เป็น Public URL

## โครงสร้างโปรเจกต์

```text
.
├── Docker-compose.yaml
├── kong.yml
├── README.md
└── user-service/
    ├── Dockerfile
    ├── go.mod
    ├── go.sum
    └── main.go
```

## สิ่งที่ต้องติดตั้ง

- Docker Desktop และ Docker Compose
- ngrok account และ Authtoken
- `curl` หรือ Postman สำหรับทดสอบ API
- Go 1.26 ขึ้นไป หากต้องการรัน User Service โดยไม่ใช้ Docker

## ตั้งค่า ngrok

สร้างไฟล์ `.env` ในโฟลเดอร์รากของโปรเจกต์ โดยไม่ต้อง commit ไฟล์นี้:

```env
NGROK_AUTHTOKEN=ใส่_ngrok_authtoken_ของคุณ
```

ห้ามใส่ Authtoken จริงลงใน `Docker-compose.yaml` หรือ commit ขึ้น Git เนื่องจากเป็นข้อมูลลับ

## เริ่มระบบด้วย Docker

รันจากโฟลเดอร์รากของโปรเจกต์:

```powershell
docker compose -f Docker-compose.yaml up --build -d
```

คำสั่งนี้จะเริ่ม PostgreSQL, User Service, Kong และ ngrok พร้อมกัน

ตรวจสอบสถานะ:

```powershell
docker compose -f Docker-compose.yaml ps
```

ดู Public URL ของ ngrok:

```powershell
docker compose -f Docker-compose.yaml logs -f ngrok
```

ให้มองหาบรรทัดลักษณะนี้:

```text
started tunnel ... url=https://xxxx.ngrok-free.dev
```

หรือเปิดหน้า ngrok inspection ได้ที่:

```text
http://localhost:4040
```

หยุดระบบ:

```powershell
docker compose -f Docker-compose.yaml down
```

หยุดระบบและลบข้อมูล PostgreSQL:

```powershell
docker compose -f Docker-compose.yaml down -v
```

## พอร์ต

| บริการ | พอร์ต | รายละเอียด |
| --- | ---: | --- |
| PostgreSQL | `5432` | ฐานข้อมูล `userdb` |
| User Service | `8080` | สมัครสมาชิกและเข้าสู่ระบบ |
| Kong | `8000` | API Gateway ภายในเครื่อง |
| ngrok inspection | `4040` | ตรวจสอบ tunnel และ request |

สำหรับการใช้งานจากภายนอก ให้ใช้ Public URL ของ ngrok แทน `localhost:8000`:

```text
https://xxxx.ngrok-free.dev
```

เส้นทางการเรียกใช้งานคือ:

```text
โลกภายนอก -> ngrok -> Kong -> User Service หรือ Chaiyos
```

## API ผ่าน Kong

### สมัครสมาชิก

เรียกจากเครื่องเดียวกัน:

```powershell
curl.exe -X POST http://localhost:8000/auth/register `
  -H "Content-Type: application/json" `
  -d '{"username":"alice","password":"password123"}'
```

เรียกจากโลกภายนอก โดยแทนที่ `<NGROK_URL>` ด้วย URL ที่ได้จาก log:

```powershell
curl.exe -X POST https://<NGROK_URL>/auth/register `
  -H "Content-Type: application/json" `
  -d '{"username":"alice","password":"password123"}'
```

ผลลัพธ์สำเร็จ:

```json
{
  "message": "User registered successfully"
}
```

Kong จะตัด prefix `/auth` ก่อนส่งต่อเป็น `POST /register` ไปยัง User Service

### เข้าสู่ระบบ

```powershell
curl.exe -X POST https://<NGROK_URL>/auth/login `
  -H "Content-Type: application/json" `
  -d '{"username":"alice","password":"password123"}'
```

ผลลัพธ์จะมี JWT token:

```json
{
  "token": "<JWT_TOKEN>"
}
```

JWT มีอายุ 2 ชั่วโมง และต้องส่งใน header เมื่อต้องการเรียก Chaiyos:

```text
Authorization: Bearer <JWT_TOKEN>
```

## Chaiyos

Chaiyos ไม่ได้รันอยู่ใน Docker Compose ชุดนี้ แต่รันอยู่บนเครื่องหรือบริการภายนอก โดย URL ปลายทางถูกกำหนดใน [kong.yml](kong.yml) เช่น:

```yaml
url: https://pursuant-battered-untimed.ngrok-free.dev
```

ถ้า Chaiyos รันอยู่บนเครื่องและมีเฉพาะ `localhost:8082` ต้องเปิด ngrok สำหรับ Chaiyos เพิ่มอีกหนึ่ง tunnel จากนั้นนำ URL ใหม่ไปแก้ใน `kong.yml` แล้วโหลด Kong ใหม่:

```powershell
docker compose -f Docker-compose.yaml up -d --force-recreate kong
```

เรียก Chaiyos ผ่าน Kong:

```powershell
curl.exe -i https://<NGROK_URL>/chaiyos `
  -H "Authorization: Bearer <JWT_TOKEN>"
```

Kong จะตรวจสอบ JWT ก่อนส่งคำขอไปยัง Chaiyos หากไม่มี token หรือ token ไม่ถูกต้อง คำขอจะถูกปฏิเสธ

## การตั้งค่า JWT

ค่าปัจจุบันสำหรับการทดลองอยู่ใน [user-service/main.go](user-service/main.go) และ [kong.yml](kong.yml):

| รายการ | ค่า |
| --- | --- |
| Issuer (`iss`) | `my-issuer-key` |
| Secret | `my-super-secret-key` |
| อายุ token | 2 ชั่วโมง |
| Kong consumer | `my-client` |

ค่าดังกล่าวใช้สำหรับการทดลองเท่านั้น ควรย้าย secret ไปไว้ใน environment variable หรือ secret manager ก่อนใช้งานจริง

## รัน User Service โดยไม่ใช้ Docker

ต้องมี PostgreSQL ที่เชื่อมต่อได้ด้วยค่าต่อไปนี้:

```text
host: localhost
port: 5432
user: postgres
password: secret
database: userdb
```

จากนั้นรัน User Service:

```powershell
cd user-service
go run .
```

User Service จะเปิดที่ `http://localhost:8080`

## หมายเหตุด้านความปลอดภัยและการพัฒนา

- ไม่ควรเปิดพอร์ต `5432` และ `8080` สู่สาธารณะ ให้เปิดผ่าน Kong เป็นทางเข้าหลัก
- เปลี่ยน Authtoken ทันทีหากเคยเผยแพร่หรือ commit ขึ้น Git
- `depends_on` ไม่ได้รอให้ PostgreSQL พร้อมรับ connection อย่างสมบูรณ์
- User Service หน่วงเวลา 5 วินาทีก่อนเชื่อมต่อ PostgreSQL เหมาะสำหรับตัวอย่างนี้เท่านั้น
- รหัสผ่านถูก hash ด้วย bcrypt ก่อนบันทึกลงฐานข้อมูล
- password ของ PostgreSQL และ JWT secret เป็นค่าเพื่อการทดลองเท่านั้น
