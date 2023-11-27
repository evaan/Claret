CREATE TABLE IF NOT EXISTS "courseTimes" (
	"crn" varchar NOT NULL,
	"days" varchar NOT NULL,
	"startTime" varchar NOT NULL,
	"endTime" varchar NOT NULL,
	"location" varchar NOT NULL,
	"ignore" serial PRIMARY KEY NOT NULL
);
--> statement-breakpoint
CREATE TABLE IF NOT EXISTS "courses" (
	"crn" varchar PRIMARY KEY NOT NULL,
	"id" varchar NOT NULL,
	"name" varchar NOT NULL,
	"section" varchar NOT NULL,
	"dateRange" varchar NOT NULL,
	"type" varchar NOT NULL,
	"instructor" varchar NOT NULL
);
--> statement-breakpoint
CREATE TABLE IF NOT EXISTS "subjects" (
	"name" varchar PRIMARY KEY NOT NULL,
	"friendlyName" varchar NOT NULL
);
--> statement-breakpoint
DO $$ BEGIN
 ALTER TABLE "courseTimes" ADD CONSTRAINT "courseTimes_crn_courses_crn_fk" FOREIGN KEY ("crn") REFERENCES "courses"("crn") ON DELETE cascade ON UPDATE no action;
EXCEPTION
 WHEN duplicate_object THEN null;
END $$;
