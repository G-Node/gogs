$(function () {
    $("#commit_message").on('input', function (event) {
        if ($(this).val() === "") {
            $("#commit").prop('disabled', true)
        } else {
            $("#commit").prop('disabled', false)
        }
    })
})
