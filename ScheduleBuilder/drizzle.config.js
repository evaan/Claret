require("dotenv").config();

export default {
  schema: "/app/db/schema.ts",
  out: "./drizzle",
  driver: "pg",
  dbCredentials: {
    connectionString: process.env.DB_URL
  }
}