import type { LoaderFunctionArgs } from "@remix-run/node";
import { Await, useAsyncValue, useLoaderData } from "@remix-run/react";
import { sql } from "drizzle-orm";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "~/components/ui/accordion"
import { db } from "~/db/db.server";
import { courseTimes, courses, subjects } from "~/db/schema";

export async function loader() {
  return {courses: (await db.select().from(courses)), courseTimes: (await db.select().from(courseTimes)), subjects: (await db.select().from(subjects))};
}

export default function Index() {
  return (
    <div className="h-screen overflow-hidden">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 px-4">
        <div>
          <p className="text-red-600 text-4xl font-bold font-sen text-center">Classes</p>
          <div className="flex flex-col">
            {/* make classes scrollable but not extend page */}
            <Accordion type="single" collapsible className="w-full flex-1">
            {useLoaderData<typeof loader>().subjects.map(subject => {
              return(
                <div key={subject.name}>
                <AccordionItem value={subject.name}>
                  <AccordionTrigger>{subject.friendlyName}</AccordionTrigger>
                  <AccordionContent className="w-[98%] pl-[1%]">Test</AccordionContent>
                </AccordionItem>
              </div>
              )
            })}
            </Accordion>
          </div>
        </div>
        <div className="col-span-2">
        <p className="text-red-600 text-4xl font-bold font-sen text-center">Schedule</p>
        </div>
      </div>
    </div>

  );
}
