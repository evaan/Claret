import { pgTable, varchar, serial } from "drizzle-orm/pg-core"

export const subjects = pgTable("subjects", {
	name: varchar("name").primaryKey().notNull(),
	friendlyName: varchar("friendlyName").notNull(),
});

export const courses = pgTable("courses", {
	crn: varchar("crn").primaryKey().notNull(),
	id: varchar("id").notNull(),
	name: varchar("name").notNull(),
	section: varchar("section").notNull(),
	dateRange: varchar("dateRange").notNull(),
	type: varchar("type").notNull(),
	instructor: varchar("instructor").notNull(),
});

export const courseTimes = pgTable("courseTimes", {
	crn: varchar("crn").notNull().references(() => courses.crn, { onDelete: "cascade" } ),
	days: varchar("days").notNull(),
	startTime: varchar("startTime").notNull(),
	endTime: varchar("endTime").notNull(),
	location: varchar("location").notNull(),
	ignore: serial("ignore").primaryKey().notNull(),
});