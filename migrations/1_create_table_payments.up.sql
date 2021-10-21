
  create table "payments" (
    "id" serial primary key,
    "transaction_id" char(16) NOT NULL,
    "authorized_by" varchar(255) NOT NULL,
    "amount" INT NOT NULL,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);