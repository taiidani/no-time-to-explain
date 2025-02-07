let addMessageBtn = document.getElementById("addMessage");
let addMessageDialog = document.getElementById("addMessageDialog");
let addMessageCancelBtn = document.getElementById("addMessageCancel");

addMessageBtn.addEventListener("click", function (evt) {
  addMessageDialog.setAttribute("open", true);
});

addMessageCancelBtn.addEventListener("click", function (evt) {
  evt.preventDefault();
  evt.stopPropagation();

  addMessageDialog.getElementsByTagName("form")[0].reset();
  addMessageDialog.removeAttribute("open");
});
