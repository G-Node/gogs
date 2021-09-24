'use strict';

import { toBool } from './util/parse_util.js'

let json = JSON.parse($('#fileContent').val());
let isDmpJson = toBool($('#isDmpJson').val());

// create the editor
let container = document.getElementById("jsoneditor");
let options = {
    mode: "tree",
    onChange: function() {
        var ct = JSON.stringify(jsonEditor.get(), null, 2);
        $('#edit_area').val(ct);
        codeMirrorEditor.setValue(ct);
        console.log("changed");
    },
    onEditable: function(node) {
        switch (node.field) {
            case "schema":
                return { field: false, value: false };
            default:
                return { field: false, value: true };
        }
    },
    onError: function(error) {
        console.log("error: ", error);
    },
};
if (isDmpJson) {
    options.schema = JSON.parse($('#schema').val());
}

let jsonEditor = new JSONEditor(container, options, json);
jsonEditor.expandAll();

// set json
// var json = {{.FileContent| Str2JS}}
// jsonEditor.set(json);

// get json
json = jsonEditor.get();
