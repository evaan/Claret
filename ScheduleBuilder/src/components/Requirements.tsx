import { useId, useState } from "react";
import { Collapse } from "react-bootstrap";
import { CourseRequirements } from "../api/types";

type RequirementTable = CourseRequirements["tables"][number];

const catalogHeaders = ["And/Or", "", "Test", "Score", "Subject", "Course Number", "Level", "Grade", ""];

function isNoInformation(text: string) {
    return /^No (?:prerequisite|corequisite)(?: course)? information available\.?$/i.test(text.trim());
}

function hasRequirements(requirements?: CourseRequirements | null): requirements is CourseRequirements {
    return !!requirements && (
        (!!requirements.text.trim() && !isNoInformation(requirements.text)) ||
        requirements.tables.some(table => table.rows.some(row => row.some(cell => cell.trim())))
    );
}

function RequirementRows({ table, title }: { table: RequirementTable; title: string }) {
    // Only reformat the known Banner layout. Unrecognized tables keep their
    // original columns and order rather than guessing at the requirement logic.
    const isCatalog = table.headers.length === catalogHeaders.length &&
        table.headers.every((header, index) => header.trim().toLowerCase() === catalogHeaders[index].toLowerCase()) &&
        table.rows.every(row => row.length === catalogHeaders.length);

    if (isCatalog) {
        return <ul className="requirement-rows" aria-label={title} role="list">
            {table.rows.map((row, index) => {
                const course = [row[4], row[5]].filter(Boolean).join(" ");
                const details = [row[6], row[7] && `Minimum grade ${row[7]}`, row[3] && `Score ${row[3]}`]
                    .filter(Boolean).join(" · ");
                return <li key={index}>
                    <p className="requirement-course">
                        {row[0] && <><span className="requirement-operator">{row[0]}</span>{" "}</>}
                         {[course, row[2]].filter(Boolean).join(" · ")}
                    </p>
                    {details && <p className="requirement-note">{details}</p>}
                </li>;
            })}
        </ul>;
    }

    const width = Math.max(table.headers.length, ...table.rows.map(row => row.length));
    const columns = Array.from({ length: width }, (_, index) => index)
        .filter(column => table.rows.some(row => row[column]?.trim()));
    return <div className="requirement-table-wrapper" role="region" aria-label={title} tabIndex={0}>
        <table className="requirement-table" aria-label={title}>
            {table.headers.length > 0 && <thead><tr>
                {columns.map(column => <th scope="col" key={column}>{table.headers[column]}</th>)}
            </tr></thead>}
            <tbody>{table.rows.map((row, rowIndex) => <tr key={rowIndex}>
                {columns.map(column => <td key={column}>{row[column]}</td>)}
            </tr>)}</tbody>
        </table>
    </div>;
}

export function RequirementsDropdown({ prerequisites, corequisites }: {
    prerequisites?: CourseRequirements | null;
    corequisites?: CourseRequirements | null;
}) {
    const [expanded, setExpanded] = useState(false);
    const contentID = useId();
    const toggleID = useId();
    if (!hasRequirements(prerequisites) && !hasRequirements(corequisites)) return null;

    return <div className="requirements-dropdown">
        <button type="button" className="requirements-toggle" id={toggleID} aria-expanded={expanded} aria-controls={contentID}
            onClick={() => setExpanded(value => !value)}>
            Requirements
        </button>
        <Collapse in={expanded} timeout={220}>
            <div id={contentID} role="region" aria-labelledby={toggleID}>
                <div className="requirements-content">
                    <Requirements title="Prerequisites" requirements={prerequisites} />
                    <Requirements title="Corequisites" requirements={corequisites} />
                </div>
            </div>
        </Collapse>
    </div>;
}

export default function Requirements({ title, requirements }: {
    title: string;
    requirements?: CourseRequirements | null;
}) {
    if (!hasRequirements(requirements)) return null;

    return <section className="requirements-section">
        <h3>{title}</h3>
        {requirements.text.trim() && !isNoInformation(requirements.text) &&
            <p className="requirement-text">{requirements.text}</p>}
        {requirements.tables.filter(table => table.rows.some(row => row.some(cell => cell.trim())))
            .map((table, index) => <RequirementRows table={table} title={title} key={index} />)}
    </section>;
}
