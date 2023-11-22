"use client";

import { AppBar, Box, Button, Dialog, DialogContent, DialogTitle, Paper, Toolbar } from "@mui/material";
import { useState } from "react";

export default function Home() {
  const [dialogOpen, setDialogOpen] = useState<boolean>(false)

  return(
    <Box>
      <Box bgcolor="#181818">
        <Toolbar disableGutters sx={{paddingX:"8px"}}>
          <h1 style={{fontSize: "30px", lineHeight:"0", marginRight: "8px"}}>CLARET</h1>
          <Button variant="contained" sx={{marginX: "8px"}} onClick={() => setDialogOpen(true)}>Select Subject</Button>
          <Button variant="contained" sx={{marginX: "8px"}}>Select Course</Button>
          <p style={{flexGrow: 1}} />
          <Button variant="outlined" sx={{color: "#d10056"}} onClick={() => open("https://github.com/evaan/Claret", "_blank")?.focus()}>Github</Button>
        </Toolbar>
      </Box>
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)}>
        <DialogTitle align="center">Select Subject</DialogTitle>
        <DialogContent>
          <p>Subject Selection Here</p>
        </DialogContent>
      </Dialog>
    </Box>
  )
}