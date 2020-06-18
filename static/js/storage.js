var thisMoment = parseInt(Date.now());
for (let localStorageKey in localStorage) {

    //Проверка, что это наш URL
    if (localStorageKey.split("/").length !== 6) {continue;};

    let resp = JSON.parse(localStorage.getItem(localStorageKey));

    //Проверка, что ссылка не устарела по времени
    let expiryDate = new Date(parseInt(resp.expiry) * 1000);
    if (expiryDate < thisMoment) {
        localStorage.removeItem(resp.url);
        continue;
    };

    let upload = document.createElement("div");
    upload.className = "upload";

    let fileLabel = document.createElement("span");

    let fileLabelLink = document.createElement("a");
    fileLabelLink.href = resp.url;
    fileLabelLink.target = "_blank";
    fileLabelLink.innerHTML = resp.url.split("/").pop() + " (" + Math.ceil(parseInt(resp.size)/1024) + " kB )";;

    fileLabel.appendChild(fileLabelLink);
    upload.appendChild(fileLabel);
    let fileActions = document.createElement("div");
    fileActions.className = "right";

    let expiryLabel = document.createElement("span");
    expiryLabel.innerHTML = " до " + expiryDate.toLocaleTimeString("ru-RU").split(":").slice(0, 2).join(":") + " " + expiryDate.toLocaleDateString("ru-RU") + " ";
    expiryLabel.className = "expiry";
    fileActions.appendChild(expiryLabel);

    let deleteAction = document.createElement("span");
    deleteAction.innerHTML = "Удалить";
    deleteAction.className = "cancel";
    deleteAction.addEventListener('click', function (ev) {
        xhr = new XMLHttpRequest();
        xhr.open("DELETE", resp.url, true);
        xhr.setRequestHeader("Linx-Delete-Key", resp.delete_key);
        xhr.onreadystatechange = function (fileLabel, fileLabelLink, deleteAction, expiryLabel, resp) {
            if (xhr.readyState == 4 && xhr.status === 200) {
                let text = document.createTextNode("Удалён ");
                fileLabel.insertBefore(text, fileLabelLink);
                fileLabel.className = "deleted";
                expiryLabel.setAttribute("style","text-decoration: line-through");
                fileActions.removeChild(deleteAction);
                localStorage.removeItem(resp.url);
            }
        }.bind(null, fileLabel, fileLabelLink, deleteAction, expiryLabel, resp);
        xhr.send();
    });
    fileActions.appendChild(deleteAction);
    upload.appendChild(fileActions);
    document.querySelector("#uploads").appendChild(upload);
}