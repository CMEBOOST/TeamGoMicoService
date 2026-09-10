# Auth Project

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
