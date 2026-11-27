import React from "react";
import { Col, Row } from "react-bootstrap";

export default function MobileViews(props: { courses: React.ReactNode; schedule: React.ReactNode }) {
    const [view, setView] = React.useState<"courses" | "schedule">("courses");
    const scrollPositions = React.useRef({ courses: 0, schedule: 0 });

    function switchView(next: typeof view) {
        if (next === view) return;
        scrollPositions.current[view] = window.scrollY;
        setView(next);
    }

    function pressView(event: React.PointerEvent<HTMLButtonElement>, next: typeof view) {
        if (!event.isPrimary || event.button !== 0) return;
        switchView(next);
    }

    React.useLayoutEffect(() => {
        if (!window.matchMedia("(max-width: 767px)").matches) return;
        window.scrollTo({ top: scrollPositions.current[view], behavior: "instant" });
    }, [view]);

    // Keep the expensive calendar and course list out of tab-switch renders.
    return <>
        <Row className="g-4">
            <Col xs={12} md={8} className="order-2 mobile-view-panel" id="schedule-panel" data-mobile-active={view === "schedule"}>
                {props.schedule}
            </Col>
            <Col xs={12} md={4} className="order-1 mobile-view-panel" id="courses-panel" data-mobile-active={view === "courses"}>
                {props.courses}
            </Col>
        </Row>
        <nav className="mobile-view-switcher" aria-label="Mobile views" data-active-view={view}>
            <button type="button" aria-pressed={view === "courses"} aria-controls="courses-panel" onPointerDown={event => pressView(event, "courses")} onClick={() => switchView("courses")}>
                Courses
            </button>
            <button type="button" aria-pressed={view === "schedule"} aria-controls="schedule-panel" onPointerDown={event => pressView(event, "schedule")} onClick={() => switchView("schedule")}>
                Schedule
            </button>
        </nav>
    </>;
}
