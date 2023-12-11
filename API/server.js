import pg from "pg";
import express from "express";
import cors from "cors";
import { config } from "dotenv";
import { spawn } from "child_process";
config();

//initialize sql
const client = new pg.Client(process.env.DB_URL)
await client.connect();

//get port from .env, or default to 8080
const port = process.env.PORT || 8080;

//initialize express app
const app = express();

app.use(cors())

console.log("\x1b[95m ####   #      ###  #####  ###### #####")
console.log("#    #  #     #   # #    # #        #")
console.log("#       #     ##### #####  ####     #")
console.log("#       #     #   # #    # #        #")
console.log("#    #  #     #   # #    # #        #")
console.log(" ####   ##### #   # #    # ######   # API v0.1")
console.log("https://github.com/evaan/Claret\x1b[0m")

app.listen(port, () => {
    console.log("\x1b[32mRunning at port", port + "!\x1b[0m");
})

app.get("/", (req, res) => {
    res.send("make a usage page here")
})

//get list of subjects
app.get("/subjects", async (req, res) => {
    res.json((await client.query("SELECT * FROM subjects")).rows);
})

//search courses by id
app.get("/courses/:id", async (req, res) => {
    res.json((await client.query("SELECT * FROM courses WHERE subject = $1", [req.params.subject + "%"])).rows);
})

//get times by crn
app.get("/times/:crn", async (req, res) => {
    res.json((await client.query("SELECT * FROM times WHERE crn = $1", [req.params.crn])).rows);
})

//return all subjects, courses, and times
app.get("/all", async (req, res) => {
    let output = {}
    output["subjects"] = (await client.query("SELECT * FROM subjects")).rows;
    output["courses"] = (await client.query("SELECT * FROM courses")).rows;
    output["times"] = (await client.query("SELECT * FROM times")).rows;
    output["seatings"] = (await client.query("SELECT * FROM seatings")).rows;
    res.json(output);
})

app.get("/seating/:crn", async(req, res) => {
    let child = spawn("python3", ["../Scraper/SeatingScrape.py", req.params.crn])
    child.on("exit", async function() {
        res.json((await client.query("SELECT * FROM seatings WHERE crn = $1", [req.params.crn])).rows)
    })
})