// document.addEventListener("DOMContentLoaded",function(){
//     var button = document.getElementById("helloButton");
//     button.addEventListener("click",function(){
//         alert("Hello");
//     })
// })
$(function () {
    let host = 'http://127.0.0.1:8080'
    $('#rolad_waller').click(function(){
            $.ajax({
              url: host+"/wallet",
              type: "POST",
              success: function (response) {
                $("#inputPublic").val(response["public_key"]);
                $("#inputPrivateKey").val(response["private_key"]);
                $("#inputAddress").val(response["blockchain_address"]);
                console.info(response);
              },
              error: function (error) {
                console.error(error);
              },
            })
    })
    $('#select_vale').change(function(){
     
      host = $('#select_vale').val()
    })


    $('#load_waller').click(function(){
      let private = $("#inputPrivateKey").val();
            $.ajax({
              url: host+"/walletByPrivatekey",
              type: "POST",
              data:{
                privatekey:private
              },
              success: function (response) {
                $("#inputPublic").val(response["public_key"]);
                $("#inputPrivateKey").val(response["private_key"]);
                $("#inputAddress").val(response["blockchain_address"]);
                console.info(response);
              },
              error: function (error) {
                console.error(error);
              },
            })
    })

  });