import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from "postgres";
import * as schema from './schema';
import { config } from 'dotenv';

config()
const queryClient = postgres(process.env.DB_URL!);
export const db = drizzle(queryClient, { schema });