import { useAtom } from "jotai";
import { useState } from "react";
import { Button, Form } from "react-bootstrap";
import { filterAtom, searchQueryAtom } from "../api/atoms";
import { activeFilterCount } from "../api/functions";
import FilterModal from "./FilterModal";

export default function SearchBar() {
    const [searchQuery, setSearchQuery] = useAtom(searchQueryAtom);
    const [filters] = useAtom(filterAtom);
    const [showFilters, setShowFilters] = useState(false);
    const count = activeFilterCount(filters);

    return (
        <>
            <div className="course-search my-3">
                <Form.Control type="search" aria-label="Search courses, instructors, rooms, and more"
                    placeholder="Search courses, instructors, rooms…" value={searchQuery}
                    onChange={event => setSearchQuery(event.target.value)} />
                <Button variant="outline-primary" onClick={() => setShowFilters(true)} aria-haspopup="dialog">
                    Filters{count > 0 ? ` (${count})` : ""}
                </Button>
            </div>
            {showFilters && <FilterModal onHide={() => setShowFilters(false)} />}
        </>
    );
}
