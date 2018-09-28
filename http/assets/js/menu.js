$(function() {
    // console.log("request data")
    $.getJSON('./assets/menu.json',"", function(data, status) {
      if (status == 'success') {
          $.each(data, function(i, menu){
              html = '<li class="treeview"><a href="' + menu.controller + '"><i class="glyphicon glyphicon-book"></i><span>'+menu.name+'</span><span class="pull-right-container"><i class="fa fa-angle-left pull-right"></i></span></a><ul id="'+menu.menuID+'" class="treeview-menu"></ul></li>'
              $(".sidebar-menu").append(html)
              $.each(menu.child, function(i, cMenu){
                  var uri = "'" + cMenu.controller + "'"

                  html = '<li class="" onclick="acd()"><a href="javascript:go('+uri+')"><i class="fa fa-circle-o"></i> '+cMenu.name+'</a></li>'
                  $("#" + menu.menuID).append(html)
              })
          })
      } else {
          console.log("没有读取到本地文件: " + status)
      }
    });
});

function go(url) {
    $("#main").attr("src", url)
}

 