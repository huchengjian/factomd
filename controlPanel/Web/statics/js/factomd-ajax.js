function queryState(a,b,c){var d=new XMLHttpRequest;d.onreadystatechange=function(){4==d.readyState&&c(d.response)},d.open("GET","/factomd?item="+a+"&value="+b,!0),d.send()}function batchQueryState(a,b){var c=new XMLHttpRequest;c.onreadystatechange=function(){4==c.readyState&&b(c.response)},c.open("GET","/factomdBatch?batch="+a,!0),c.send()}function searchBarSubmit(){var a=new XMLHttpRequest;a.onreadystatechange=function(){4==a.readyState&&(console.log(a.response),obj=JSON.parse(a.response),"dblockHeight"==obj.Type?window.location="search?input="+obj.item+"&type=dblock":"None"!=obj.Type?window.location="search?input="+$("#factom-search").val()+"&type="+obj.Type:($(".factom-search-error").slideDown(300),console.log(a.response)))};var b=new FormData;b.append("method","search"),b.append("search",$("#factom-search").val()),a.open("POST","/post"),a.send(b)}function redirect(a,b,c){var d=$("<input>").attr("type","hidden").val(c).attr("name","content"),e=$("<form>",{method:b,action:a});e.append(d),e.submit()}function nextNode(){resp=queryState("nextNode","",function(a){$("#current-node-number").text(a)})}$("#factom-search").click(function(){$(".factom-search-error").slideUp(300)}),$("#factom-search-submit").click(function(){searchBarSubmit()}),$(".factom-search-container").keypress(function(a){var b=a.which||a.keyCode;13==b&&searchBarSubmit()}),$("body").on("mouseup","section #factom-search-link",function(a){type=jQuery(this).attr("type"),hash=jQuery(this).text();var b=new XMLHttpRequest;b.onreadystatechange=function(){4==b.readyState&&(obj=JSON.parse(b.response),"None"!=obj.Type?1==a.which?window.location="search?input="+hash+"&type="+type:2==a.which&&window.open("/search?input="+hash+"&type="+type):"special-action-fack"==obj.Type?window.location="search?input="+hash+"&type="+type:$(".factom-search-error").slideDown(300))};var c=new FormData;c.append("method","search"),c.append("search",hash),c.append("known",type),b.open("POST","/post"),b.send(c)});