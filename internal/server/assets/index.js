document.getElementById("addMessage").addEventListener("click", function (evt) {
    document.getElementById("addMessageDialog").setAttribute("open", true)
})

document.getElementById("addMessageCancel").addEventListener("click", function (evt) {
    evt.preventDefault()
    evt.stopPropagation()

    document.getElementById("addMessageForm").reset()
    document.getElementById("addMessageDialog").removeAttribute("open")
})
