# Auth Project

ตัวอย่างระบบยืนยันตัวตนด้วย Go ประกอบด้วย User Service, PostgreSQL และ Kong API Gateway สำหรับเรียกใช้บริการ Chaiyos โดย User Service ทำหน้าที่สมัครสมาชิก เข้าสู่ระบบ และออก JWT

## ภาพรวมระบบ

```text
Client
  |
  v
Kong :8000
  |-- /auth/*  -> user-service:8080
  |-- /chaiyos -> Chaiyos ภายนอกผ่าน ngrok (ต้องใช้ JWT)
  |
PostgreSQL :5432
```

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
- `curl` หรือ Postman สำหรับทดสอบ API
- ngrok สำหรับเปิด Chaiyos ให้ Kong เข้าถึง
- Go 1.26 ขึ้นไป หากต้องการรัน User Service โดยไม่ใช้ Docker

## เริ่มระบบด้วย Docker

รันจากโฟลเดอร์รากของโปรเจกต์:

```powershell
docker compose -f Docker-compose.yaml up --build -d
```

ตรวจสอบสถานะ:

```powershell
docker compose -f Docker-compose.yaml ps
```

ดู log:

```powershell
docker compose -f Docker-compose.yaml logs -f user-service
docker compose -f Docker-compose.yaml logs -f kong
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
| Kong | `8000` | API Gateway |

## API ผ่าน Kong

Base URL:

```text
http://localhost:8000
```

### สมัครสมาชิก

```powershell
curl.exe -X POST http://localhost:8000/auth/register `
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
curl.exe -X POST http://localhost:8000/auth/login `
  -H "Content-Type: application/json" `
  -d '{"username":"alice","password":"password123"}'
```

ผลลัพธ์จะมี JWT token:

```json
{
  "token": "<JWT_TOKEN>"
}
```

JWT มีอายุ 2 ชั่วโมง ใช้ token ใน header เมื่อต้องการเรียก Chaiyos:

```text
Authorization: Bearer <JWT_TOKEN>
```

## Chaiyos

Chaiyos ไม่ได้รันอยู่ใน Docker Compose ชุดนี้ แต่รันอยู่บนเครื่องหรือบริการภายนอก และเปิดให้ Kong เข้าถึงผ่าน ngrok เช่น:

```text
https://pursuant-battered-untimed.ngrok-free.dev -> http://localhost:8082
```

ใน [kong.yml](kong.yml) route `/chaiyos` จะส่งต่อไปยัง Chaiyos และใช้ JWT plugin ตรวจสอบ token:

```yaml
- name: chaiyos-service
  url: https://pursuant-battered-untimed.ngrok-free.dev
  routes:
    - name: chaiyos-route
      paths:
        - /chaiyos
      strip_path: false
```

### เรียก Chaiyos ผ่าน Kong

```powershell
curl.exe -i http://localhost:8000/chaiyos `
  -H "Authorization: Bearer <JWT_TOKEN>"
```

Kong จะตรวจสอบ JWT ก่อนส่งคำขอไปยัง Chaiyos หากไม่มี token หรือ token ไม่ถูกต้อง คำขอจะถูกปฏิเสธ

หาก ngrok URL เปลี่ยน ให้แก้ค่า `url` ใน [kong.yml](kong.yml) แล้วโหลด Kong ใหม่:

```powershell
docker compose -f Docker-compose.yaml up -d --force-recreate kong
```

ทดสอบทั้งปลายทาง Chaiyos และ Kong ได้ด้วยคำสั่ง:

```powershell
curl.exe -i https://pursuant-battered-untimed.ngrok-free.dev/chaiyos `
  -H "Authorization: Bearer <JWT_TOKEN>"

curl.exe -i http://localhost:8000/chaiyos `
  -H "Authorization: Bearer <JWT_TOKEN>"
```

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

## หมายเหตุ

- `depends_on` ไม่ได้รอให้ PostgreSQL พร้อมรับ connection อย่างสมบูรณ์
- User Service หน่วงเวลา 5 วินาทีก่อนเชื่อมต่อ PostgreSQL เหมาะสำหรับตัวอย่างนี้เท่านั้น
- รหัสผ่านถูก hash ด้วย bcrypt ก่อนบันทึกลงฐานข้อมูล
- password ของ PostgreSQL และ JWT secret เป็นค่าเพื่อการทดลองเท่านั้น
