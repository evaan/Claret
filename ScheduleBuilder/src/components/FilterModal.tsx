import { useState } from "react";
import { useAtom, useAtomValue } from "jotai";
import { Button, Col, Form, Modal, Row } from "react-bootstrap";
import { coursesAtom, filterAtom } from "../api/atoms";
import { campuses, defaultFilters } from "../api/functions";
import { CourseFilters } from "../api/types";

const days = [["M", "Mon"], ["T", "Tue"], ["W", "Wed"], ["R", "Thu"], ["F", "Fri"], ["S", "Sat"], ["U", "Sun"]];

export default function FilterModal(props: { onHide: () => void }) {
    const [show, setShow] = useState(true);
    const [filters, setFilters] = useAtom(filterAtom);
    const courses = useAtomValue(coursesAtom);
    const [draft, setDraft] = useState<CourseFilters>(filters);
    const types = [...new Set([...courses.map(course => course.type), ...draft.types])].filter(Boolean).sort();
    const delivery = [...new Set([...courses.map(course => course.instructionalMethod ?? ""), ...draft.delivery])].filter(Boolean).sort();
    const invalidTime = !!(draft.startsAfter && draft.endsBefore && draft.startsAfter >= draft.endsBefore);
    const invalidCredits = draft.minCredits !== "" && draft.maxCredits !== "" && Number(draft.minCredits) > Number(draft.maxCredits);
    const invalid = invalidTime || invalidCredits;

    function update<K extends keyof CourseFilters>(key: K, value: CourseFilters[K]) {
        setDraft(previous => ({ ...previous, [key]: value }));
    }

    function toggle(key: "campuses" | "days", value: string) {
        setDraft(previous => ({
            ...previous,
            [key]: previous[key].includes(value) ? previous[key].filter(item => item !== value) : [...previous[key], value],
        }));
    }

    return <Modal show={show} onHide={() => setShow(false)} onExited={props.onHide} centered scrollable>
        <Modal.Header closeButton><Modal.Title>Filters</Modal.Title></Modal.Header>
        <Form className="filter-form" onSubmit={event => {
            event.preventDefault();
            if (invalid) return;
            setFilters(draft);
            setShow(false);
        }}>
            <Modal.Body>
                <fieldset className="filter-group">
                    <legend>Campus</legend>
                    <div className="filter-options">
                        {campuses.map((campus, index) => <Form.Check key={campus} id={`filter-campus-${index}`} label={campus}
                            checked={draft.campuses.includes(campus)} onChange={() => toggle("campuses", campus)} />)}
                    </div>
                </fieldset>
                <Row className="g-3 mb-3">
                    <Col xs={6}>
                        <Form.Label htmlFor="filter-type">Section type</Form.Label>
                        <Form.Select id="filter-type" value={draft.types[0] ?? ""} onChange={event => update("types", event.target.value ? [event.target.value] : [])}>
                            <option value="">Any</option>
                            {types.map(type => <option key={type}>{type}</option>)}
                        </Form.Select>
                    </Col>
                    <Col xs={6}>
                        <Form.Label htmlFor="filter-delivery">Delivery</Form.Label>
                        <Form.Select id="filter-delivery" value={draft.delivery[0] ?? ""} onChange={event => update("delivery", event.target.value ? [event.target.value] : [])}>
                            <option value="">Any</option>
                            {delivery.map(method => <option key={method}>{method}</option>)}
                        </Form.Select>
                    </Col>
                </Row>
                <fieldset className="filter-group">
                    <legend>Meeting days</legend>
                    <div className="filter-options">
                        {days.map(([value, label]) => <Form.Check key={value} id={`filter-day-${value}`} label={label}
                            checked={draft.days.includes(value)} onChange={() => toggle("days", value)} />)}
                    </div>
                    <p className="detail-muted">Any selected day. Leave blank for all.</p>
                </fieldset>
                <Row className="g-3 mb-3">
                    <Col xs={6}>
                        <Form.Label htmlFor="filter-start">Starts at or after</Form.Label>
                        <Form.Control id="filter-start" type="time" value={draft.startsAfter} isInvalid={invalidTime}
                            onChange={event => update("startsAfter", event.target.value)} />
                    </Col>
                    <Col xs={6}>
                        <Form.Label htmlFor="filter-end">Ends at or before</Form.Label>
                        <Form.Control id="filter-end" type="time" value={draft.endsBefore} isInvalid={invalidTime}
                            onChange={event => update("endsBefore", event.target.value)} />
                    </Col>
                    <Col xs={12} className="mt-1">
                        <p className="detail-muted">{invalidTime ? "End time must be after start time." : "All scheduled meetings must fit. Unscheduled sections are excluded when a time is set."}</p>
                    </Col>
                    <Col xs={6}>
                        <Form.Label htmlFor="filter-min-credits">Min. credits</Form.Label>
                        <Form.Control id="filter-min-credits" type="number" min="0" step="any" placeholder="Any"
                            value={draft.minCredits} isInvalid={invalidCredits} onChange={event => update("minCredits", event.target.value)} />
                    </Col>
                    <Col xs={6}>
                        <Form.Label htmlFor="filter-max-credits">Max. credits</Form.Label>
                        <Form.Control id="filter-max-credits" type="number" min="0" step="any" placeholder="Any"
                            value={draft.maxCredits} isInvalid={invalidCredits} onChange={event => update("maxCredits", event.target.value)} />
                    </Col>
                    {invalidCredits && <Col xs={12} className="mt-1"><p className="detail-muted">Maximum credits must be at least the minimum.</p></Col>}
                </Row>
                <Form.Check id="filter-available" label="Available seats only" checked={draft.availableOnly}
                    onChange={event => update("availableOnly", event.target.checked)} />
            </Modal.Body>
            <Modal.Footer>
                <Button variant="link" className="me-auto" onClick={() => setDraft(defaultFilters())}>Reset</Button>
                <Button variant="secondary" onClick={() => setShow(false)}>Cancel</Button>
                <Button type="submit" disabled={invalid}>Apply filters</Button>
            </Modal.Footer>
        </Form>
    </Modal>;
}
