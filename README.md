## mgo for php

MongoDB ConnPool using hprose-golang + mgo.v2 in this project. 
Maybe you can use it to resolve the problem "PHP-mongodb extention can not close connection".

## How to use it?

Use it at PHP script like this:

```
<?php
require_once "hprose-php/src/Hprose.php";

use Hprose\Client;

try{
    $client = Client::create('http://127.0.0.1:8080/', false);
    
    $params= [
        'database'=> 'mall',
        'collection'=> 'tag_goods_sales_rank',
        //'select'=> '+gid,+logo,+qty,-_id',
        //'sort'=> '-show,-sale,+_id',
        //'offset'=> 1,
        //'limit'=> 2,
    ];
    $where= [
        'name'=>'/粤通卡/',
        //'gid'=> 111,
        //'qty'=> 1,
        //'sale'=> "1",
    ];
    header("Content-type: text/html; charset=utf-8");
    echo $client->getData($params, $where);  //you can define more and more function at server-side
    
} catch (Exception $e){
    echo $e->getMessage();
}

```
