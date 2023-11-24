"use client";

import { Accordion, AccordionDetails, AccordionSummary, Box, Button, Grid, Toolbar, Typography } from "@mui/material";

export default function Home() {
  return(
    <Box sx={{height: "100vh"}}>
      <Box bgcolor="#181818">
        <Toolbar disableGutters sx={{paddingX:"8px"}}>
          <a href="/"><Button sx={{color: "#d10056"}}><h1 style={{fontSize: "30px", lineHeight:"0", marginRight: "8px"}}>CLARET</h1></Button></a>
          <p style={{flexGrow: 1}} />
          <a href="https://github.com/evaan/Claret" target="_blank"><Button variant="outlined" sx={{color: "#d10056"}}>Github</Button></a>
        </Toolbar>
      </Box>
      <Grid container spacing={1} paddingX={"8px"}>
        <Grid item xs={12} sm={4}>
          {/* would be cool to make the course list not extend the page on desktop */}
          <h1 style={{textAlign: "center"}}>Courses</h1>
          <Accordion>
            <AccordionSummary><Typography align="center" sx={{width: "100%"}}>Computer Science</Typography></AccordionSummary>
            <AccordionDetails>
              <Accordion sx={{background: "#282828"}}>
                <AccordionSummary><Typography align="center" sx={{width: "100%"}}>COMP 1001</Typography></AccordionSummary>
                <AccordionDetails>
                  <Box display="flex" flexWrap="wrap">
                    <Button variant="contained" sx={{marginY: "4px", flexBasis: "100%"}}>Section 001 - Lecture (time here maybe)</Button>
                    <Button variant="contained" sx={{marginY: "4px", flexBasis: "100%"}}>Section 002 - Laboratory (time here maybe)</Button>
                    <Button variant="contained" sx={{marginY: "4px", flexBasis: "100%"}}>Section 003 - Laboratory (time here maybe)</Button>
                  </Box>
                </AccordionDetails>
              </Accordion>
            </AccordionDetails>
          </Accordion>
          <Accordion>
            <AccordionSummary><Typography align="center" sx={{width: "100%"}}>Mathematics</Typography></AccordionSummary>
            <AccordionDetails>
              <Accordion sx={{background: "#282828"}}>
                <AccordionSummary><Typography align="center" sx={{width: "100%"}}>MATH 1000</Typography></AccordionSummary>
                <AccordionDetails>
                  <Box display="flex" flexWrap="wrap">
                    <Button variant="contained" sx={{marginY: "4px", flexBasis: "100%"}}>Section 001 - Lecture (time here maybe)</Button>
                    <Button variant="contained" sx={{marginY: "4px", flexBasis: "100%"}}>Section 002 - Lecture (time here maybe)</Button>
                    <Button variant="contained" sx={{marginY: "4px", flexBasis: "100%"}}>Section 003 - Lecture (time here maybe)</Button>
                  </Box>
                </AccordionDetails>
              </Accordion>
            </AccordionDetails>
          </Accordion>
        </Grid>
        <Grid item xs={12} sm={8}>
          <h1 style={{textAlign: "center"}}>Schedule</h1>
        </Grid>
      </Grid>
    </Box>
  )
}