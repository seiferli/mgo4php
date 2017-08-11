<?php
require_once "../../../hprose-php/src/Hprose.php";

use Hprose\Client;
use Hprose\TimeoutException;

header("Content-type: text/html; charset=utf-8");

try{
    $client = Client::create('http://127.0.0.1:8080/', false);
    $params= [
        'database'=> 'mall',
        'collection'=> 'tag_goods_rank',
        //'select'=> '+gid,+logo,+qty,-_id', //- means un-select the field,
        //'sort'=> '-show,-sale,+_id',  //+: asc sorting -: desc sorting
        //'offset'=> 100,
        'limit'=> 20,
    ];
    $where= [
        "or",
        [
            "and",
            ['>', '_id', 1],
            ['<=', '_id', 100],
        ],
        ["%", "title", "iphone"],
    ];
    echo $client->getData($params, []);  //you can define more and more function at server-side
    
} catch (Exception $e){
    echo $e->getMessage();
}
