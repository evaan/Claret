import { Button, Center, Select, Text } from "@chakra-ui/react";
import { CategoryScale } from "chart.js";
import Chart from "chart.js/auto";
import { Line } from 'react-chartjs-2';

Chart.register(CategoryScale);

export default function App() {
    const data = {
        labels: ['Red', 'Orange', 'Blue'],
        datasets: [
            {
              label: 'Popularity of colours',
              data: [55, 23, 96],
              borderColor: '#81E6D9',
              backgroundColor: '#81E6D9',
              borderWidth: 1,
            }
        ]
    };

    return (
        <Center height="100vh" width="100vw" overflowX="hidden">
            <div style={{width: "90%"}}>
                <Text fontSize='6xl' align="center" marginBottom="16px">Claret Seat Monitor</Text>
                <Select placeholder='Select engineering discipline' size="lg" marginBottom="16px" width="90%" marginX="5%">
                    <option value='engi'>Engineering (ENGI)</option>
                    <option value='civ'>Civil Engineering (CIV)</option>
                    <option value='ece'>Electrical/Computer Engineering (ECE)</option>
                    <option value='me'>Mechanical Engineering (ME)</option>
                    <option value='onae'>Ocean/Naval Engineering (ONAE)</option>
                    <option value='proc'>Process Engineering (PROC)</option>
                </Select>
                <Select placeholder='Select course' size="lg" marginBottom="16px" width="90%" marginX="5%">
                    <option value='1010'>ENGI 1010 (Engineering Statics)</option>
                </Select>
                <Button colorScheme="teal" marginBottom="16px" width="90%" marginX="5%">View Graph</Button>
                <div style={{width: "100%", height: "50vh"}}>
                    <Line data={data} options={{maintainAspectRatio: false}} width="100%" height="100%" />
                </div>
            </div>
        </Center>
    )
}