# 03 - Podatkovni model

## Entitete in glavna polja

### 1. `User` (Uporabnik)
* `id` (PK, UUID)
* `email` (String, Unique)
* `password_hash` (String)
* `first_name` (String)
* `last_name` (String)
* `role` (Enum: `EMPLOYEE`, `OFFICE_MANAGER`, `ADMIN`)
* `department` (String)
* `created_at` (Timestamp)

### 2. `Resource` (Vir: Miza, Sejna soba, Parkirišče)
* `id` (PK, UUID)
* `name` (String, npr. "Miza B12", "Sejna soba Alfa")
* `type` (Enum: `DESK`, `MEETING_ROOM`, `PARKING_SPOT`)
* `floor` (Integer)
* `capacity` (Integer - število oseb, relevantno za sobe)
* `is_active` (Boolean)
* `created_at` (Timestamp)

### 3. `ResourceFeature` (Oprema / Lastnost vira)
* `id` (PK, UUID)
* `resource_id` (FK -> `Resource.id`)
* `feature_name` (Enum/String: `DUAL_MONITOR`, `ADJUSTABLE_DESK`, `DOCKING_STATION`, `PROJECTOR`, `QUIET_ZONE`, `WHITEBOARD`)

### 4. `Reservation` (Rezervacija)
* `id` (PK, UUID)
* `user_id` (FK -> `User.id`)
* `resource_id` (FK -> `Resource.id`)
* `start_time` (Timestamp)
* `end_time` (Timestamp)
* `status` (Enum: `CONFIRMED`, `CHECKED_IN`, `COMPLETED`, `CANCELLED`, `EXPIRED`)
* `check_in_code` (String, 4-mestna koda)
* `checked_in_at` (Timestamp, Nullable)
* `created_at` (Timestamp)

---

## Povezave med entitetami
* **User – Reservation:** $1 : N$ (En uporabnik lahko v času ustvari več rezervacij).
* **Resource – Reservation:** $1 : N$ (Posamezen vir ima lahko več rezervacij v različnih časovnih intervalih).
* **Resource – ResourceFeature:** $1 : N$ (En vir ima lahko določenih več različnih kosov opreme ali lastnosti).

---

## Diagram podatkovnega modela (Mermaid)

![Podatkovni model DeskBooker](images/database_diagram.png)

```mermaid
erDiagram
    USER ||--o{ RESERVATION : "creates"
    RESOURCE ||--o{ RESERVATION : "is booked in"
    RESOURCE ||--o{ RESOURCE_FEATURE : "has"

    USER {
        uuid id PK
        string email
        string password_hash
        string first_name
        string last_name
        string role
        string department
        timestamp created_at
    }

    RESOURCE {
        uuid id PK
        string name
        string type
        int floor
        int capacity
        boolean is_active
        timestamp created_at
    }

    RESOURCE_FEATURE {
        uuid id PK
        uuid resource_id FK
        string feature_name
    }

    RESERVATION {
        uuid id PK
        uuid user_id FK
        uuid resource_id FK
        timestamp start_time
        timestamp end_time
        string status
        string check_in_code
        timestamp checked_in_at
        timestamp created_at
    }