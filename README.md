## mgo for php (mgo4php)

> MongoDB Connetion Pool using hprose-golang + mgo.v2 .  Maybe you can use it to resolve the problem "PHP-mongodb extention can not close connection".

## How to use it on client side (such as php)?

> **step1: Install Git, and Download the project source using GO GET command tools**
```
cd /home/www/gopath/src/
go get gopkg.in/mgo.v2
go get github.com/valyala/fasthttp
go get github.com/hprose/hprose-golang
go get github.com/hprose/hprose-php  ## Make sure your php version > 5.5
go get github.com/seiferli/mgo4php
```
> **step2: Make the mongodb configuration file**
```
cd /home/www/gopath/src/github.com/seiferli/mgo4php
cp config.ini.sample config.ini
vi config.ini  ## Input your setting paramters. 
```
> **step3: Server Side： Running Mgo4php service**
```
cd /home/www/gopath/src/github.com/seiferli/mgo4php
go run main.go # you can modify listen port at main.go
```
> **step4: Client Side： run the demo script**
```
cd /home/www/gopath/src/github.com/seiferli/mgo4php/mgo
php client.php # or visit script output on web browser
```

## What can we do at client.php througth the mgo4php ?

> **Basic calling**
```
<?php
require_once "../../../hprose-php/src/Hprose.php";

use Hprose\Client;
use Hprose\TimeoutException;

header("Content-type: text/html; charset=utf-8");

$client = Client::create('http://127.0.0.1:8080/', false);
$params= [
    'database'=> '[dbname]',
    'collection'=> '[collection]',
];

// running server side function
echo $client->one($params, []); # return one matching record
#
### return the result json
{
    code: 0,
    msg: "ok",
    data: {
        _id: 1,
        del: 0,
        gid: 790,
        logo: "https://image.etcchebao.com/20160314182500cjXwqz.jpg",
        name: "商品BBB1",
    }
}

```

> **Read simple data**
```
$params= [
    'database'=> '[dbname]',            # not options! 
    'collection'=> '[collection]',      # not options! 
    
    //"-" means un-select the field, unset the key when return all field
    'select'=> '+gid,+logo,+qty,-_id',  # options 
    
    //+: asc sorting -: desc sorting
    'sort'=> '-show,-sale,+_id',        # options 
    
    'offset'=> 100, // offset number
    'limit'=> 20,  // limit count       # options 
];

// return all matching record
echo $client->all($params, []);  
#

// return all matching count
echo $client->count($params, []);  
#

// return all matching count
echo $client->all($params, $where= [
    'type'=> "hot",
    'sale'=> "1",     // string type
    'status'=> 1,     // int type
]);  
#

```

> **Complex query expression**
```

# compare ">" "<" ">=" "<=" "!" "%"
$where= [ ">", "_id", 11 ];
$where= [ ">=", "_id", 11 ];
$where= [ "<", "_id", 99 ];
$where= [ "<=", "_id", 99 ];
$where= [ "=", "_id", 55 ];
$where= [ "!", "_id", 77 ];
$where= [ "%", "title", "someword" ];
#
# "in" condition
$where= [ "in", "_id", [ 1,3,4,5 ] ];
#
# "and" condition
$where= [
    "and",
    [ ">=", "_id", 11 ],
    [ "<", "_id", 99 ],
    ...
];
#
# "or" condition
$where= [
    "or",
    [
        "and",
        [ ">=", "_id", 11 ],
        [ "<", "_id", 99 ],
        ...
    ],
    [
        'qty'=> "<100",
        'sale'=> "1", 
    ],
    [ "%", "title", "somestring"],
    ...
];
#
```


> **Insert / update / batchInsert / delete**
```
# base insert 

$params= [
    'database'=> '[dbname]',            # not options! 
    'collection'=> '[collection]',      # not options! 
];

// simple insert
echo $client->insert($params, ["title"=>"test", "status"=>1, "content"=>"some content"... ]);  
#
#
# nesting insert bson data
echo $client->insert($params, [
     "string"=>"hello world", "_id"=> 123, "array"=>[ 
         's1'=> "ssss", 
         's2'=> [ 
            1, 3, 34  
         ]
     ],
 ]);
 
#
# delete bson data
echo $client->delete($params, $where= ["_id"=> 123] );

#
# update the rows
echo $client->update($params, $where= ["_id"=> 123], [
[
    "date"=> date('Y-m-d H:i:s'), 
    "arr"=> [
        "string"=> "new string", 
    ]
] );

#
# rewrite the rows
echo $client->update($params, $where= ["_id"=> 123], [
    'reflesh', 
    [
        "date"=>date('Y-m-d H:i:s'), 
        "arr"=> [
            "string"=> "new string", 
            "arr"=> [ 1, 2, 3] 
        ] 
    ] 
] );
    
#
# insert batch
echo $client->batchInsert($params, [
    [ 
        "string"=>"fffffffff", "array"=> [ 
            's1'=> "ssss", 's2'=> 999999999 
        ]
    ],
    [ 
        's1'=> "ssss", "array"=>[ 
            's1'=> "ssss", 
            's2'=> [ 
                's1'=> "ssss", 's2'=> 999999999 
            ] 
        ] 
    ],
    ...
] );
    
```

> **More advance syntax analysis**

```
develop...


```
