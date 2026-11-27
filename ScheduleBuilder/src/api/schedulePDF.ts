import { jsPDF } from "jspdf";
import moment from "moment";
import { buildSchedule, ScheduleEvent } from "./schedule";
import { formatTime } from "./functions";
import { Course, Time } from "./types";

export function positionPDFEvents(layout: ReturnType<typeof buildSchedule>) {
    const positioned: (ScheduleEvent & { column: number; lane: number; lanes: number })[] = [];
    layout.days.forEach((_, column) => {
        const events = layout.events.filter(event =>
            moment(event.start).startOf("day").diff(moment(layout.range.start), "days") === column
        ).sort((a, b) => a.start.localeCompare(b.start) || a.end.localeCompare(b.end));
        let cluster: ScheduleEvent[] = [];
        let clusterEnd = "";
        const flush = () => {
            const laneEnds: string[] = [];
            const assigned = cluster.map(event => {
                let lane = laneEnds.findIndex(end => end <= event.start);
                if (lane === -1) lane = laneEnds.length;
                laneEnds[lane] = event.end;
                return { ...event, column, lane };
            });
            positioned.push(...assigned.map(event => ({ ...event, lanes: laneEnds.length })));
            cluster = [];
        };
        events.forEach(event => {
            if (cluster.length && event.start >= clusterEnd) flush();
            if (!cluster.length) clusterEnd = event.end;
            else if (event.end > clusterEnd) clusterEnd = event.end;
            cluster.push(event);
        });
        flush();
    });
    return positioned;
}

export function createSchedulePDF(courses: Course[], times: Time[], semester: string) {
    const layout = buildSchedule(courses, times);
    const pdf = new jsPDF({ orientation: "landscape", unit: "pt", format: "letter", compress: true, putOnlyUsedFonts: true });
    pdf.setProperties({ title: `${semester} schedule`, author: "Claret" });
    pdf.setDisplayMode("fullwidth", "single");
    pdf.viewerPreferences({ PrintScaling: "None", DisplayDocTitle: true });
    const margin = 30;
    const tableTop = margin + 24;
    const headerHeight = 24;
    const axisWidth = 43;
    const width = pdf.internal.pageSize.getWidth() - margin * 2;
    const height = pdf.internal.pageSize.getHeight() - margin - tableTop;
    const gridTop = tableTop + headerHeight;
    const gridHeight = height - headerHeight;
    const columnWidth = (width - axisWidth) / layout.days.length;
    const spanMinutes = (layout.max - layout.min) * 60;
    const yForMinutes = (minutes: number) => gridTop + (minutes - layout.min * 60) / spanMinutes * gridHeight;
    const minutesForDate = (date: string) => {
        const time = moment(date);
        return time.hour() * 60 + time.minute();
    };

    pdf.setFont("helvetica", "normal");
    pdf.setFontSize(11);
    pdf.text(semester, margin, margin + 10);
    pdf.setFontSize(9);
    pdf.setFillColor(245, 245, 245);
    pdf.rect(margin, tableTop, width, headerHeight, "F");
    pdf.setDrawColor(170);
    pdf.setLineWidth(0.4);
    pdf.rect(margin, tableTop, width, height);
    pdf.line(margin, gridTop, margin + width, gridTop);
    layout.days.forEach((day, index) => {
        const x = margin + axisWidth + index * columnWidth;
        pdf.line(x, tableTop, x, tableTop + height);
        pdf.text(day, x + columnWidth / 2, tableTop + 15, { align: "center" });
    });
    const slotMinutes = layout.max - layout.min > 12 ? 30 : 15;
    for (let minutes = layout.min * 60; minutes <= layout.max * 60; minutes += slotMinutes) {
        const y = yForMinutes(minutes);
        pdf.setDrawColor(minutes % 60 === 0 ? 170 : 220);
        pdf.setLineWidth(minutes % 60 === 0 ? 0.4 : 0.2);
        pdf.line(margin + axisWidth, y, margin + width, y);
        if (minutes % (layout.max - layout.min > 12 ? 60 : 30) === 0) {
            pdf.setFontSize(7);
            pdf.text(formatTime(`${Math.floor(minutes / 60)}:${String(minutes % 60).padStart(2, "0")}`),
                margin + axisWidth - 4, y + (minutes === layout.max * 60 ? -2 : 7), { align: "right" });
        }
    }

    positionPDFEvents(layout).forEach(event => {
        // Match the calendar's 1px margins, with square corners and side-by-side conflicts.
        const laneWidth = columnWidth / event.lanes;
        const eventWidth = laneWidth - 1.5;
        const x = margin + axisWidth + event.column * columnWidth + event.lane * laneWidth + 0.75;
        const y = yForMinutes(minutesForDate(event.start));
        const eventHeight = (minutesForDate(event.end) - minutesForDate(event.start)) / spanMinutes * gridHeight;
        if (eventHeight <= 0) return;
        pdf.setDrawColor(85);
        pdf.setLineWidth(0.5);
        pdf.setFillColor(255, 255, 255);
        pdf.rect(x, y, eventWidth, eventHeight, "FD");
        const time = event.extendedProps?.unscheduled ? "" :
            `${formatTime(event.start.slice(11))} - ${formatTime(event.end.slice(11))}`;
        let fontSize = 8;
        const linesForSize = () => {
            pdf.setFontSize(fontSize);
            return pdf.splitTextToSize([time, event.title].filter(Boolean).join("\n"), eventWidth - 6) as string[];
        };
        let lines = linesForSize();
        while (fontSize > 6 && lines.length * fontSize * 1.15 > eventHeight - 6) {
            fontSize -= 0.5;
            lines = linesForSize();
        }
        pdf.saveGraphicsState();
        pdf.rect(x + 1, y + 1, eventWidth - 2, Math.max(0, eventHeight - 2), null);
        pdf.clip();
        pdf.discardPath();
        pdf.setTextColor(0);
        pdf.text(lines, x + 3, y + 3 + fontSize, { lineHeightFactor: 1.15 });
        pdf.restoreGraphicsState();
    });
    return pdf;
}
