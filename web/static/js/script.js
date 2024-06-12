const elementForm = document.getElementById("form")
const elementMessage = document.getElementById("message")
const elementSubmit = document.getElementById("submit")

async function sendData() {
    const textInput = document.getElementById("text")
    const fileInput = document.getElementById("file")

    const formData = new FormData();
    const encoder = new TextEncoder();
    const contentBytes = encoder.encode(textInput.value);

    if (contentBytes.length > 1000) {
        alert('Textarea content exceeds 1000 bytes limit.');
        return;
    }

    formData.append('file', fileInput.files[0]);
    formData.append('text', textInput.value);

    try {
        const response = await fetch("/api/content", {
            method: "POST",
            body: formData
        })

        if (response.ok) {
            console.log('File and content uploaded successfully!');
        } else {
            console.error('Upload failed.');
        }

        console.debug("Response:")
        console.log(response)

        const body = await response.json()
        if (response.status === 400) {
            await createError(body.message)
            return
        }

        if (response.status > 499 && response.status < 600) {
            await createError(body)
            return
        }

        const message = window.location.protocol + "//" + window.location.host + "/?key=" + body.message
        await createMessage(message)
        elementSubmit.disabled = true
    } catch (e) {
        console.error(e)
        await createError(e)
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
    sendData().then(r => {console.debug("Sent!")})
});
