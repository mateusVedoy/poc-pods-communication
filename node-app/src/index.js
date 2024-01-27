const express = require("express")
const os = require("os")

const app = express()

const router = express.Router()

host = os.hostname()

router.post("/", (_, res) => {

    console.log(`${host} received the message`)

    res.status(200).send({ "message": `OK from pod ${host}` })
})

app.use(express.json())

app.use("/node-app", router)

const server = app.listen("8080", () => {
    console.log(`${host} is running on port 8080`)
})

process.on("SIGTERM", shutdown);
process.on("SIGINT", shutdown);

function shutdown() {
    console.log(`${host} is shutting down`)
    server.close()
    process.exit(0)
}