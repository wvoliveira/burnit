const elementForm = document.getElementById("form")
const elementMessage = document.getElementById("message")
const elementSubmit = document.getElementById("submit")

async function sendData() {
    const data = document.getElementById("secret")

    try {
        const response = await fetch("/", {
            method: "POST",
            body: data.value,
        })
        console.debug("Response:")
        console.log(response)

        const body = await response.json()
        if (response.status == 400) {
            createError(body.message)
            return
        }

        if (response.status > 499 && response.status < 600) {
            createError(body)
            return
        }

        const message = window.location.protocol + "//" + window.location.host + "/?key=" + body.message
        createMessage(message)
        elementSubmit.disabled = true
    } catch (e) {
        console.error(e)
        createError(e)
    }
}

async function createError(error) {
    const elementParagraph = document.createElement("p");
    elementParagraph.textContent = "Error: " + error
    elementMessage.append(elementParagraph)
}

async function createMessage(message) {
    const elementParagraph = document.createElement("p");
    elementParagraph.textContent = message

    elementMessage.innerHTML = ""
    elementMessage.append(elementParagraph)
}

elementForm.addEventListener("submit", (event) => {
    event.preventDefault()
    sendData()
});
