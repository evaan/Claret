"use client";

import { ThemeProvider } from "@emotion/react";
import { CssBaseline, createTheme } from "@mui/material";
import { Sen } from 'next/font/google'

const inter = Sen({subsets: ['latin']})

const theme = createTheme({
    palette: {
        mode: 'dark',
        background: {
            default: "#121212",
            paper: "#000000"
        },
        primary: {
            main: '#7f1734',
        },
        text: {
            primary: "#d10056"
        }
    },
    typography: {
        fontFamily: inter.style.fontFamily,
    }
});

export default function Providers({children}: {children: React.ReactNode}) {
    return(
        <ThemeProvider theme={theme}>
            <CssBaseline />
            {children}
        </ThemeProvider>
    )
}