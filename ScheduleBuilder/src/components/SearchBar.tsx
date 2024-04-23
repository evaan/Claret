import { useAtom } from "jotai";
import { Form } from "react-bootstrap";
import { searchQueryAtom } from "../api/atoms";

export default function SearchBar() {
    const [, setSearchQuery] = useAtom(searchQueryAtom);

    return (
        <Form.Control type="textarea" placeholder="Search for courses" className="my-2" onChange={(event) => setSearchQuery(event.target.value)}/>
    );
}