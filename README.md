# Auth Project

ตัวอย่างระบบ Microservices ด้วย Go โดยมี User Service, Student Service, PostgreSQL และ Kong API Gateway สำหรับรวม endpoint และตรวจสอบ JWT

## โครงสร้างระบบ

```text
Client
  |
  v
Kong :8000
  |-- /auth/*    -> user-service:8080
  |-- /students  -> student-service:8081
  |-- /chaiyos   -> Chaiyos ภายนอกผ่าน ngrok
  |
PostgreSQL :5432
```

Chaiyos ไม่ได้รันอยู่ใน Docker Compose ชุดนี้ แต่รันอยู่บนเครื่องอื่นและเปิดให้เข้าถึงผ่าน ngrok เช่น:

```text
https://pursuant-battered-untimed.ngrok-free.dev
```

## สิ่งที่ต้องติดตั้ง

- Docker Desktop และ Docker Compose
- Postman หรือ `curl`
- ngrok เฉพาะกรณีที่ต้องเปิด Chaiyos จากเครื่องอื่น

## เริ่มระบบ

รันจากโฟลเดอร์รากของโปรเจกต์:

```powershell
docker compose -f Docker-compose.yaml up --build -d
```

ตรวจสอบสถานะ:

```powershell
docker compose -f Docker-compose.yaml ps
```

หยุดระบบ:

```powershell
docker compose -f Docker-compose.yaml down
```

หยุดและลบข้อมูล PostgreSQL ด้วย:

```powershell
docker compose -f Docker-compose.yaml down -v
```

## พอร์ต

| บริการ | พอร์ต | รายละเอียด |
| --- | ---: | --- |
| PostgreSQL | `5432` | ฐานข้อมูล `userdb` |
| User Service | `8080` | สมัครสมาชิกและ login |
| Student Service | `8081` | จัดการข้อมูลนักศึกษา |
| Kong | `8000` | API Gateway |
| Chaiyos | `8082` บนเครื่องอื่น | เข้าผ่าน ngrok ไม่ได้อยู่ใน Compose |

## API ผ่าน Kong

Base URL ในเครื่อง:

```text
http://localhost:8000
```

ถ้าเปิด Kong ด้วย ngrok ให้ใช้ URL สาธารณะของ Kong แทน เช่น:

```text
https://platter-handrail-thrash.ngrok-free.dev
```

### สมัครสมาชิก

```http
POST /auth/register
Content-Type: application/json
```

```json
{
  "username": "alice",
  "password": "password123"
}
```

PowerShell:

```powershell
curl.exe -X POST http://localhost:8000/auth/register `
  -H "Content-Type: application/json" `
  -d '{"username":"alice","password":"password123"}'
```

Kong จะตัด prefix `/auth` ก่อนส่งต่อเป็น `POST /register` ไปยัง User Service

### Login

```http
POST /auth/login
Content-Type: application/json
```

```json
{
  "username": "alice",
  "password": "password123"
}
```

PowerShell:

```powershell
curl.exe -X POST http://localhost:8000/auth/login `
  -H "Content-Type: application/json" `
  -d '{"username":"alice","password":"password123"}'
```

ผลลัพธ์จะมี token:

```json
{
  "token": "<JWT_TOKEN>"
}
```

### เพิ่มนักศึกษา

```http
POST /students
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

```json
{
  "name": "Alice Smith",
  "student_id": "6500001"
}
```

### ดูนักศึกษาทั้งหมด

```http
GET /students
Authorization: Bearer <JWT_TOKEN>
```

ทั้งสอง endpoint ของ `/students` ต้องใช้ JWT ที่ได้จาก `/auth/login`

### Chaiyos

```http
GET /chaiyos
```

ตัวอย่างเรียกผ่าน Kong:

```powershell
curl.exe http://localhost:8000/chaiyos
```

ผลลัพธ์ตัวอย่าง:

```json
{
  "name": "ชัยยศ ศรีสว่าง",
  "student_id": "67114540088"
}
```

## การตั้งค่า Chaiyos

เครื่องที่รัน Chaiyos ต้องเปิด endpoint ให้ ngrok forward ไปยัง port `8082`:

```text
https://pursuant-battered-untimed.ngrok-free.dev -> http://localhost:8082
```

ใน `kong.yml` ต้องกำหนด URL ของ Chaiyos ให้ตรงกับ ngrok URL ปัจจุบัน:

```yaml
- name: chaiyos-service
  url: https://pursuant-battered-untimed.ngrok-free.dev
  routes:
    - name: chaiyos-route
      paths:
        - /chaiyos
      strip_path: false
```

หลังเปลี่ยน URL หรือแก้ `kong.yml` ให้โหลด config ใหม่:

```powershell
docker compose -f Docker-compose.yaml up -d --force-recreate kong
```

ทดสอบจากเครื่องที่รัน Kong ก่อน:

```powershell
curl.exe -i https://pursuant-battered-untimed.ngrok-free.dev/chaiyos
curl.exe -i http://localhost:8000/chaiyos
```

ทั้งสองคำสั่งควรได้ `200 OK`

## การตั้งค่า JWT

| รายการ | ค่าในตัวอย่าง |
| --- | --- |
| Issuer (`iss`) | `my-issuer-key` |
| Secret | `my-super-secret-key` |
| อายุ token | 2 ชั่วโมง |
| Kong consumer | `my-client` |

ค่าดังกล่าวอยู่ใน `user-service/main.go` และ `kong.yml` ใช้สำหรับการทดลองเท่านั้น ควรย้าย secret ไปไว้ใน environment variable หรือ secret manager ก่อนใช้งานจริง

## หมายเหตุสำคัญเกี่ยวกับ Student Service

ใน `kong.yml` service ของ route `/students` ควรชี้ไปที่ container `student-service`:

```yaml
- name: student-service
  url: http://student-service:8081
```

ถ้าพบว่า `/students` เชื่อมต่อผิด service ให้ตรวจค่าดังกล่าว แล้ว recreate Kong:

```powershell
docker compose -f Docker-compose.yaml up -d --force-recreate kong
```

## ดู log

```powershell
docker compose -f Docker-compose.yaml logs -f user-service
docker compose -f Docker-compose.yaml logs -f student-service
docker compose -f Docker-compose.yaml logs -f kong
```

## การพัฒนาโดยไม่ใช้ Docker

PostgreSQL ต้องเข้าถึงได้ด้วยค่า:

```text
host: localhost
port: 5432
user: postgres
password: secret
database: userdb
```

รัน User Service:

```powershell
cd user-service
go run .
```

รัน Student Service ในอีก terminal:

```powershell
cd student-service
go run .
```

โค้ดจะรอ 5 วินาทีก่อนเชื่อมต่อ PostgreSQL เพื่อให้ฐานข้อมูลเริ่มทำงาน

## ข้อควรระวัง

- password ของ PostgreSQL และ JWT secret เป็นค่าเพื่อการทดลองเท่านั้น
- `depends_on` ไม่ได้รอให้ PostgreSQL พร้อมรับ connection อย่างสมบูรณ์
- ควรเพิ่ม health check, retry และการจัดการ error ก่อนใช้งานจริง
- การเรียก Student Service ที่ port `8081` โดยตรงจะไม่ผ่าน JWT validation ของ Kong

โปรเจกต์ตัวอย่างระบบ Microservices ด้วย Go ประกอบด้วย User Service, Student Service, PostgreSQL และ Kong API Gateway สำหรับตรวจสอบ JWT

## โครงสร้างโปรเจกต์

```text
.
├── Docker-compose.yaml
├── kong.yml
├── user-service/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── chaiyos-service/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── main.go
└── student-service/
    ├── Dockerfile
    ├── go.mod
    ├── go.sum
    └── main.go
```

## ความสามารถ

- สมัครสมาชิกและเข้าสู่ระบบ
- เก็บรหัสผ่านเป็น bcrypt hash
- สร้าง JWT เมื่อเข้าสู่ระบบสำเร็จ โดย token มีอายุ 2 ชั่วโมง
- จัดการข้อมูลนักศึกษา
- ใช้ PostgreSQL เป็นฐานข้อมูลร่วม
- ใช้ Kong เป็น API Gateway และตรวจสอบ JWT สำหรับ Student Service

## สิ่งที่ต้องติดตั้ง

- Docker
- Docker Compose
- `curl` หรือ Postman สำหรับทดสอบ API

## การเริ่มระบบ

รันคำสั่งจากโฟลเดอร์รากของโปรเจกต์:

```bash
docker compose -f Docker-compose.yaml up --build
```

บริการจะเปิดพอร์ตดังนี้:

| บริการ | พอร์ต | รายละเอียด |
| --- | ---: | --- |
| PostgreSQL | `5432` | ฐานข้อมูล `userdb` |
| User Service | `8080` | สมัครสมาชิกและเข้าสู่ระบบ |
| Student Service | `8081` | จัดการข้อมูลนักศึกษา |
| Chaiyos Service | `8082` | ข้อมูล mock ของชัยยศ |
| Kong | `8000` | API Gateway |

หยุดระบบด้วยคำสั่ง:

```bash
docker compose -f Docker-compose.yaml down
```

ลบข้อมูล PostgreSQL ที่ค้างอยู่ด้วยคำสั่งเพิ่มเติม:

```bash
docker compose -f Docker-compose.yaml down -v
```

## API Endpoints

### User Service

เรียกตรงผ่านพอร์ต `8080` หรือเรียกผ่าน Kong ด้วย prefix `/auth`

#### สมัครสมาชิก

```http
POST http://localhost:8000/auth/register
Content-Type: application/json
```

```json
{
  "username": "alice",
  "password": "password123"
}
```

ตัวอย่างด้วย `curl`:

```bash
curl -X POST http://localhost:8000/auth/register ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"alice\",\"password\":\"password123\"}"
```

#### เข้าสู่ระบบ

```http
POST http://localhost:8000/auth/login
Content-Type: application/json
```

```json
{
  "username": "alice",
  "password": "password123"
}
```

ตัวอย่างด้วย `curl`:

```bash
curl -X POST http://localhost:8000/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"alice\",\"password\":\"password123\"}"
```

ผลลัพธ์จะมี JWT token:

```json
{
  "token": "<JWT_TOKEN>"
}
```

### Student Service

#### เพิ่มข้อมูลนักศึกษา

Student Service ต้องส่ง JWT ใน header:

```http
POST http://localhost:8000/students
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

```json
{
  "name": "Alice Smith",
  "student_id": "6500001"
}
```

ตัวอย่างด้วย `curl`:

```bash
curl -X POST http://localhost:8000/students ^
  -H "Authorization: Bearer <JWT_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -d "{\"name\":\"Alice Smith\",\"student_id\":\"6500001\"}"
```

#### ดึงข้อมูลนักศึกษาทั้งหมด

```bash
curl http://localhost:8000/students ^
  -H "Authorization: Bearer <JWT_TOKEN>"
```

หากเรียก Student Service โดยตรง สามารถใช้ URL เหล่านี้ได้:

```text
POST http://localhost:8081/students
GET  http://localhost:8081/students
```

การเรียกตรงจะไม่ผ่านการตรวจสอบ JWT ของ Kong

### Chaiyos Service

ดึงข้อมูล mock ของชัยยศโดยไม่เชื่อมต่อฐานข้อมูล:

```http
GET http://localhost:8000/chaiyos
```

หรือเรียก service โดยตรง:

```http
GET http://localhost:8082/chaiyos
```

ผลลัพธ์:

```json
{
  "name": "ชัยยศ ศรีสว่าง",
  "student_id": "67114540088"
}
```

## การตั้งค่า JWT

ค่าที่ใช้ในโปรเจกต์ปัจจุบัน:

| รายการ | ค่า |
| --- | --- |
| JWT issuer (`iss`) | `my-issuer-key` |
| JWT secret | `my-super-secret-key` |
| อายุ token | 2 ชั่วโมง |
| Kong consumer | `my-client` |

ค่าดังกล่าวอยู่ใน `user-service/main.go` และ `kong.yml` สำหรับการใช้งานจริงควรย้าย secret ไปไว้ใน environment variable หรือ secret manager

## หมายเหตุเกี่ยวกับ Kong

ใน `kong.yml` ปัจจุบัน service ของ route `/students` ถูกตั้งค่าให้ส่งต่อไปยัง URL ภายนอกนี้:

```text
https://platter-handrail-thrash.ngrok-free.dev/
```

ดังนั้นการเรียก `http://localhost:8000/students` จะไม่ส่งต่อไปยัง container `student-service` โดยตรง หากต้องการให้ Kong เรียก Student Service ที่รันใน Docker Compose ให้เปลี่ยนเป็น:

```yaml
- name: student-service
  url: http://student-service:8081
```

และคง route `/students` ไว้ตามเดิม จากนั้น restart ด้วยคำสั่ง:

```bash
docker compose -f Docker-compose.yaml up --build
```

## การดู log

```bash
docker compose -f Docker-compose.yaml logs -f user-service
docker compose -f Docker-compose.yaml logs -f student-service
docker compose -f Docker-compose.yaml logs -f kong
```

## การพัฒนาโดยไม่ใช้ Docker

ต้องมี PostgreSQL ที่เข้าถึงได้ด้วยค่าต่อไปนี้:

```text
host: localhost
port: 5432
user: postgres
password: secret
database: userdb
```

รัน User Service:

```bash
cd user-service
go run .
```

รัน Student Service ในอีก terminal:

```bash
cd student-service
go run .
```

หมายเหตุ: โค้ดจะหน่วงเวลา 5 วินาทีก่อนเชื่อมต่อฐานข้อมูลเพื่อรอ PostgreSQL เริ่มทำงาน

## ข้อควรระวัง

- ค่า password และ JWT secret ในไฟล์ปัจจุบันเป็นค่าเพื่อการทดลองเท่านั้น
- ควรตรวจสอบ error จาก `AutoMigrate` และ `r.Run` ก่อนใช้งานจริง
- ควรเพิ่ม health check และระบบ retry สำหรับการเชื่อมต่อ PostgreSQL
- ควรเพิ่มการตรวจสอบสิทธิ์และ validation ให้ละเอียดขึ้นสำหรับ production
