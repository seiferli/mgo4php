<?php
require_once "hprose-php/src/Hprose.php";

use Hprose\Client;
use Hprose\TimeoutException;

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
        "and",
        [
            'qty'=> 1,
        ],
        [
            'sale'=> "1", 
        ],
    ];
    header("Content-type: text/html; charset=utf-8");
    echo $client->getData($params, $where);  //you can define more and more function at server-side
    
} catch (Exception $e){
    echo $e->getMessage();
}

/*
```
    $where= [
        //'name'=> '/someword/',  //like "%someword%" at mysql
        //'qty'=> 1,
        //'sale'=> "1",     //attention at the value type: int? or string?
    ];

#The above is the basic format, you can define it more complex
;
# compare ">" "<" "!"
$where= [
    'qty'=> ">10",
    'id'=> "<10",
    'id'=> "=10",  //=10
    'id'=> ">=10",  //not 10
    'id'=> "<=10",  //not 10
    'id'=> "!10",  //not 10
];
;
# "in" condition
$where= [
    "in",
    [ 1,3,4,5 ],
    ...
];
;
# "and" condition
$where= [
    "and",
    [
        'qty'=> 1,
    ],
    [
        'sale'=> "1", 
    ],
    ...
];
;
# "or" condition
$where= [
    "or",
    [
        'qty'=> ">10",
        'sale'=> "1", 
    ],
    [
        'qty'=> "<100",
        'sale'=> "1", 
    ],
    ...
];
;
# and this... 
$where= [
    "and",
    [
        "or",
        [
            'qty'=> ">10",
            'sale'=> "1", 
        ],
        [
            'qty'=> "<100",
            'sale'=> "1", 
        ],
    ],
    [
        'del'=> 1,
    ],
    [
        "in",
        [ 1,3,4,5 ],
    ]
    ...
];
*/
