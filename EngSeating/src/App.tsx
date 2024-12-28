import { CategoryScale } from "chart.js";
import Chart from "chart.js/auto";
import { useState } from "react";
import { Form } from "react-bootstrap";
import { Line } from 'react-chartjs-2';

Chart.register(CategoryScale);

export default function App() {
    const [discipline, setDiscipline] = useState<string>("Select Discipline");
    const [course, setCourse] = useState<string>("Select Course");
    
    const data = {
        labels: ['Red', 'Orange', 'Blue'],
        datasets: [
            {
              label: "Registered Seats",
              data: [55, 23, 96],
              borderColor: '#A8415B',
              backgroundColor: '#A8415B',
              borderWidth: 1,
            }
        ]
    };
    return (
        <div style={{display: "flex", height: "100vh", width: "100vw", alignItems: "center", justifyContent: "center", overflowX: "hidden"}}>
            <div style={{width: "90%", height: "80%"}}>
                <h1 style={{textAlign: "center", marginBottom: "16px"}}>Engineering Seat Monitor</h1>
                <Form.Select size="lg" style={{marginBottom: "16px", width: "90%", marginLeft: "5%", marginRight: "5%"}} onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {setDiscipline(e.target.value); setCourse("Select Course");}} isInvalid={discipline === "Select Discipline"} value={discipline}>
                    <option>Select Discipline</option>
                    <option>Civil Engineering</option>
                    <option>Engineering Systems</option>
                    <option>Engineering Graphics</option>
                    <option>Electrical/Computer Engineer</option>
                    <option>Engineering</option>
                    <option>Process Engineering</option>
                    <option>Ocean/Naval Engineering</option>
                    <option>Mechivanical Engineering</option>
                </Form.Select>
                <Form.Select size="lg" style={{marginBottom: "16px", width: "90%", marginLeft: "5%", marginRight: "5%"}} onChange={(e: React.ChangeEvent<HTMLSelectElement>) => {setCourse(e.target.value)}} isInvalid={course === "Select Course"} disabled={discipline === "Select Discipline"} value={course}>
                    <option>Select Course</option>
                    <option>ENGI 1010</option>
                    <option>ECE 3400</option>
                    <option>ENGI 3101</option>
                </Form.Select>
                {(course != "Select Course" && discipline != "Select Dicipline") &&
                    <div style={{width: "100%", height: "50vh"}}>
                        <Line data={data} options={{maintainAspectRatio: false}} width="100%" height="100%" />
                    </div>
                }

            </div>
        </div>
    )
}
