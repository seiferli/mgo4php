<?php
// Copyright © 2017 seiferli <469997798@qq.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

namespace BaseComponents\base;

/**
 * mongodb数据处理客户端
 * @package BaseComponents\base
 * @author  469997798@qq.com
 *
 * 使用方法：
 * 如yii2中使用，在config 编辑配置文件，components 节点中增加
 *  'mongoproxy' => [
 *     'class' => 'BaseComponents\base\Mongoproxy',
 *  ],
 * 后，即可直接调用 \yii::$app->mongoproxy->setDatabase('mall')->findAll('order');
 */
class Mongoproxy extends \yii\base\Component
{
    //查询指令 operators

    //更新指令 operators
    const UPDATE_OP_NONE  = 0;
    const UPDATE_OP_SET   = 1;
    const UPDATE_OP_UNSET = 2;
    const UPDATE_OP_PUSH  = 3;
    const UPDATE_OP_PUSHA = 4;
    const UPDATE_OP_PULL  = 5;
    const UPDATE_OP_PULLA = 6;
    const UPDATE_OP_POP   = 7;
    const UPDATE_OP_INC   = 8;
    const UPDATE_OP_ATSET = 9;
    const UPDATE_OP_RENAM = 10;

    protected $version= '1.0.1';

    protected static $client= null; //
    public $database= null;
    protected $isDebug= false;

    public function getClient()
    {
        //暂无法支持多帐号切换

        header("Content-type: text/html; charset=utf-8");
        if( self::$client==null){
            require(VDIR . '/framework/vendor/hprose-php/src/Hprose.php');
            return \Hprose\Client::create('http://api.mongoproxy.com/', false);
        }
        return self::$client;
    }

    //通过切换版本调用不同内置方法
    public function setVersion($version)
    {
        $this->version= $version;
        return $this;
    }

    public function setDebug($isDebug)
    {
        if($isDebug){
            $this->isDebug= true;
        }
        return $this;
    }

    public function setDatabase($dbname)
    {
        if($dbname){
            $this->database= $dbname;
        }
        return $this;
    }

    public function getDatabase()
    {
        if($this->database==null){
            throw new \Exception('必须先用setDatabase指定所操作数据库！');
        } else {
            return $this->database;
        }
    }

    /**
     * 构建查询过滤条件
     * @param $collection
     * @param array $select  eg: [] or ['name', 'price',...] or ['name'=>'+', 'status'=>'-']
     * @param array $sort   eg: ['name', 'price',...] or ['name'=>'desc', 'status'=>'asc']
     * @param int $offset
     * @param bool $limit
     * @return array
     */
    protected function _buildQueryCondition($collection, $select=[], $sort=[], $offset=1, $limit=false)
    {
        $condition= [
            'database'=> $this->getDatabase(),
            'collection'=> $collection,
            //'select'=> '+gid,+logo,+qty,-_id', //- means un-select the field,
            //'sort'=> '-show,-sale,+_id',  //+: asc sorting -: desc sorting
            //'offset'=> 100,
            //'limit'=> 20,
        ];
        if(is_string($select)){
            $select= explode(',', $select);
        }
        if(count($select)>0){
            $tmp= [];
            foreach ($select as $k=>$v){
                //值为- 时代表隐藏该字段
                if($v=='-' || $v=='+'){
                    $tmp[]= ($v=='-')? "-$k": "+$k";
                } else {
                    $tmp[]= "+$v";
                }
            }
            $condition['select'] = implode(',', $tmp);
        }
        if(is_string($sort)){
            $sort= explode(',', $sort);
        }
        if(count($sort)>0){
            $tmp= [];
            foreach ($sort as $k=>$v){
                //值为- 时代表隐藏该字段
                if($v=='desc' || $v=='asc'){
                    $tmp[]= ($v=='desc')? "-$k": "+$k";
                } else {
                    $tmp[]= "-$v";  //默认为倒数
                }
            }
            $condition['sort'] = implode(',', $tmp);
        }
        if($offset>1)
            $condition['offset'] = intval($offset);
        if( $limit )
            $condition['limit'] = intval($limit);

        if($this->isDebug){
            CoreHelper::openLog(['[condition]'=> $condition], ['mongoproxy']);
        }
        return $condition;
    }

    protected function _buildQueryFilter($filter)
    {
        /** 请根据实际情况组装出以下格式的过滤条件
        # compare ">" "<" ">=" "<=" "!" "%"
        $filter= [ ">", "_id", 11 ];
        $filter= [ ">=", "_id", 11 ];
        $filter= [ "<", "_id", 99 ];
        $filter= [ "<=", "_id", 99 ];
        $filter= [ "=", "_id", 55 ];
        $filter= [ "!", "_id", 77 ];
        $filter= [ "%", "title", "someword" ];

        # "in" condition
        $filter= [ "in", "_id", [ 1,3,4,5 ] ];

        # "and" condition
        $filter= [
            "and",
            [ ">=", "_id", 11 ],
            [ "<", "_id", 99 ],
            ...
        ];

        # "or" condition
        $filter= [
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
         */
        if($this->isDebug){
            CoreHelper::openLog(['[filter]'=> $filter], ['mongoproxy']);
        }
        return $filter;
    }

    protected function _output($string)
    {
        if($this->isDebug){
            CoreHelper::openLog('[output] :'. str_replace("\n", "", $string), ['mongoproxy']);
        }
        return json_decode($string, true);
    }

    //根据条件拉取所有数据
    public function findAll($collection, $filter=[], $select=[], $sort=[], $offset=1, $limit=false)
    {
        $client= $this->getClient();
        $condition= $this->_buildQueryCondition($collection, $select, $sort, $offset, $limit);
        return $this->_output( $client->all($condition, $this->_buildQueryFilter($filter)) );
    }

    //根据条件拉取第一条数据
    public function findOne($collection, $filter=[], $select=[], $sort=[], $offset=1, $limit=false)
    {
        $client= $this->getClient();
        $condition= $this->_buildQueryCondition($collection, $select, $sort, $offset, $limit);
        return $this->_output( $client->one($condition, $this->_buildQueryFilter($filter)) );
    }

    //计算数据行数
    public function count($collection, $filter=[] )
    {
        $client= $this->getClient();
        $condition= $this->_buildQueryCondition($collection );
        return $this->_output( $client->count($condition, $this->_buildQueryFilter($filter)) );
    }

    //删除数据，注意传入正确的过滤条件
    public function delete($collection, $filter=false)
    {
        if($filter==false || $filter==null || $filter==[] ){
            throw new \Exception('您正在执行全表删除，如需全部清空请传入"all"字符');
        } elseif($filter=='all') {
            $filter= [];
        }
        $client= $this->getClient();
        $condition= $this->_buildQueryCondition($collection );
        return $this->_output( $client->delete($condition, $this->_buildQueryFilter($filter)) );
    }

    //插入单条数据，支持多层结构
    public function insert($collection, $data=[])
    {
        $client= $this->getClient();
        $condition= $this->_buildQueryCondition($collection );
        return $this->_output( $client->insert($condition, json_encode($data) ) );
    }
    //批量插入数据，支持多层结构
    public function batchInsert($collection, $dataGroup=[])
    {
        $client= $this->getClient();
        $condition= $this->_buildQueryCondition($collection );
        return $this->_output( $client->batchInsert($condition, json_encode($dataGroup) ) );
    }

    /**
     *  mongodb指令参数，具体用法参考文章：@see  https://www.cnblogs.com/yu-zhang/p/5210966.html
     *
     * $inc ：increment the match field value.		eg: {$unset:{field:N}}
     * $set ：modify the match field value			eg: {$unset:{field:S}}
     * $unset ：delete the match field value			eg: {$unset:{field:1}}
     * $push ：add one element into the match field  				eg: {$push:{"field":"elementN"}}
     * $pushAll ：add element into the field value(more then one)	eg: {$pushAll:{"field":["e1","e2"]}}
     * $addToSet : like $push if the value not exist				eg: {$push:{"field":"Michael"}}
     * $pull ：delete the match field value							eg: {$pull:{"field":"elementN"}}
     * $pullAll ：delete the match field value(more then one)		eg: {$pullAll:{"field":["e1","e2"]}}
     * $pop : delete first element:-1, or last element:1			eg: {$pop:{"field":-1/1}}
     * $rename : change the field name								eg: {$rename:{"field1":"field2"}}
     */
    protected function _buildUpdateData(array $data, int $operators )
    {
        switch ($operators ){
            case self::UPDATE_OP_SET:
                $data= ['$set'=> $data]; break;
            case self::UPDATE_OP_UNSET:
                $data= ['$unset'=> $data]; break;
            case self::UPDATE_OP_PUSH:
                $data= ['$push'=> $data]; break;
            case self::UPDATE_OP_PUSHA:
                $data= ['$pushAll'=> $data]; break;
            case self::UPDATE_OP_PULL:
                $data= ['$pull'=> $data]; break;
            case self::UPDATE_OP_PULLA:
                $data= ['$pullAll'=> $data]; break;
            case self::UPDATE_OP_POP:
                $data= ['$pop'=> $data]; break;
            case self::UPDATE_OP_INC:
                $data= ['$inc'=> $data]; break;
            case self::UPDATE_OP_ATSET:
                $data= ['$addToSet'=> $data]; break;
            case self::UPDATE_OP_RENAM:
                $data= ['$rename'=> $data]; break;
        }
        return $data;
    }

    /**
     * 更新表中部分数据，$data 默认传入修改信息  $isUpsert为true时，不存在则自动插入数据
     * @param $collection
     * @param array $filter
     * @param array $data
     * @param int $updateOperators  默认为'$set'操作指令，也可以通过传入0，再在$data中传入原生的BSON数组来进行复杂的更新指令
     * @param bool $isUpsert    $isUpsert为true时，不存在则自动插入数据
     * @param bool $multi      $multi为false时，只修改第一条匹配的数据，为true时修改全表数据
     * @return array
     */
    public function update(string $collection, array $filter=[], array $data=[], int $updateOperators=1, bool $isUpsert=false, bool $multi=true)
    {
        $client= $this->getClient();
        $condition= $this->_buildQueryCondition($collection );
        /**
         * 注意：如要传入过滤条件 _id 为Object类型，需要传入 _id_ ，举例:
         * ->update('collection', ['_id_'=>'59dc3c78c550fe50365e903f'], [ 'k1'=>'8989', 'k2'=>'v99']);
         */
        if( $updateOperators>0 ){
            $data= $this->_buildUpdateData($data, $updateOperators);
        }
        if( $updateOperators==self::UPDATE_OP_NONE  && $multi== true ){
            //mongodb要求$multi=true时，更新表达式中$update必须使用如$set等指令，避免错误修改全部数据
            $multi= false;
        }
        return $this->_output( $client->update($condition, json_encode($filter), json_encode($data), $isUpsert, $multi) );
    }

    /**
     * 直接传入原生$query BSON 和 $update BSON，进行update
     * @param string $collection
     * @param BSON $query   eg：{'$and': [{'$or':[{'price':0.99 },{'price':1.99}]}, {'$or':[{'sale':true},{'qty ':20}]} ] }
     * @param BSON $update  eg：{ $inc: { <field1>: <amount1>, <field2>: <amount2>, ... } }
     * @param bool $isUpsert   $isUpsert为true时，不存在则自动插入数据
     * @param bool $multi      $multi为false时，只修改第一条匹配的数据，为true时修改全表数据
     * @return array
     */
    public function updateBson(string $collection, string $query, string $update, bool $isUpsert=false, bool $multi=true)
    {
        $client= $this->getClient();
        $condition= $this->_buildQueryCondition($collection );
        if( $multi== true && strpos($update, '$')===false ){
            //mongodb要求$multi=true时，更新表达式中$update必须使用如$set等指令，避免错误修改全部数据
            $multi= false;
        }
        return $this->_output( $client->update($condition, $query, $update, $isUpsert, $multi) );
    }

}