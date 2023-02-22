$(document).ready(function() {
    $(".navbar-burger").click(function() {
        $(".navbar-burger").toggleClass("is-active");
        $(".navbar-menu").toggleClass("is-active");
    })
})

function onCopy(e) {
    var text = $(e).parent().prev().text();
    var $temp = $("<input>");
    $("body").append($temp);
    $temp.val(text).select();
    document.execCommand("Copy");
    $temp.remove();
    $("body").append('<div class="notification has-text-primary" id="msg"><i>✔</i><p>复制成功</p></div>')
    this.timer = setTimeout(function() {
        $(".notification ").toggleClass('show')
    }, 50)
    this.timer = setTimeout(function() {
        $(".notification ").toggleClass('show')
        this.time = setTimeout(function(){
            $(".notification ").remove();
        }, 500)
    }, 2000)
}

function onCopyThis(e) {
    var text = $(e).data("copy");
    var $temp = $("<input>");
    $("body").append($temp);
    $temp.val(text).select();
    document.execCommand("Copy");
    $temp.remove();
    $("body").append('<div class="notification has-text-primary" id="msg"><i>✔</i><p>复制成功</p></div>')
    this.timer = setTimeout(function() {
        $(".notification ").toggleClass('show')
    }, 50)
    this.timer = setTimeout(function() {
        $(".notification ").toggleClass('show')
        this.time = setTimeout(function(){
            $(".notification ").remove();
        }, 500)
    }, 2000)
}