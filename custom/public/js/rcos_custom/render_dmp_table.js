// createTable is RCOS specific code.
// This generates a table to check the information of the DMP
// created by the user and displays it on the screen.
function createTable(element, items) {
    for (const item in items) {
        let tr = document.createElement("tr");

        // add key
        let field = document.createElement("td");
        field.innerHTML = item
        tr.appendChild(field);

        // add value
        if (typeof items[item] === "object") {
            createTable(tr, items[item])
        } else {
            let value = document.createElement("td");
            value.innerHTML = items[item]
            tr.appendChild(value);
        }
        element.appendChild(tr);
    }
}

$(document).ready(function () {
    let tableEle = document.getElementById("dmp");
    let items = JSON.parse($('#items').val());
    (document.getElementById("title")).innerHTML = "dmp.json (" + items.schema + ")";
    createTable(tableEle, items)
});
