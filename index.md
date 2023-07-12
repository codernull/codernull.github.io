

## Redis设计与实现 (数据库技术丛书) 1
### 第一部分“数据结构与对象”
### 第二部分“单机数据库的实现”
### 第三部分“多机数据库的实现”
## Redis 实战
# 数据结构
## Redis 常见数据类型和应用场景

  - <span style="color:#a85d45;border-bottom:2px solid #a85d45;">1.Redis 常见数据类型和应用场景</span>
	  -   **Redis 五种数据类型的应用场景**:
		1. **String** 类型的应用场景:缓存对象、常规计数、分布式锁、共享session信息等。
		2. **List** 类型的应用场景:消息队列 <span><div class="border1">(  有两个问题:  <br>    1. 生产者需要自行实现全局唯一 ID;<br>2. 不能以消费组形式消费数据)等。</div></span>
		3. **Hash** 类型:缓存对象、购物⻋等。 
		4. **Set** 类型:聚合计算(并集、交集、差集)场景,比如点赞、共同关注、抽奖活动等。 
		5. **Zset** 类型:排序场景,比如:<b style="background:linear-gradient(to right, #9932CC, #1E90FF);background-clip:text;-webkit-background-clip:text;color:transparent">排行榜</b>、电话和姓名排序等。
	-  **Redis 后续版本又支持四种数据类型,它们的应用场景如下:** 
		1. **BitMap**(*2.2 版新增*):二值状态统计的场景,比如签到、判断用户登陆状态、连续签到用户总数等;
		2. **HyperLogLog**(*2.8 版新增*):海量数据基数统计的场景,比如百万级网⻚ UV 计数等; 
		3. **GEO**(*3.2 版新增*):存储地理位置信息的场景,比如滴滴叫⻋; 
		4. **Stream**(*5.0 版新增*):消息队列,相比于基于 List 类型实现的消息队列,有这两个特有的特性:自动生成全局唯一消息ID,支持以消费组形式消费数据。
	- 针对 Redis 是否适合做消息队列,关键看你的业务场景: 如果你的业务场景足够简单,对于数据丢失不敏感,而且**消息积压概率比较小**的情况下,把 Redis 当作队列是完全可以的。 
	- 如果你的业务有海量消息,消息积压的概率比较大,并且不能接受数据丢失,那么还是用专业的消息队列中间件吧。[](marginnote3app://note/15AF07A9-CAC6-4EBD-9F6D-0609F6B1B70C)
	 -  **String(字符串),Hash(哈希),List (列表),Set(集合)、Zset(有序集合) 。**[](marginnote3app://note/ED636ECE-C506-4B49-9573-7C6DBFFBAA47)
		-  <span class = "high2">String</span>[](marginnote3app://note/44CD82C5-E040-404D-BF21-177A488DEE40)
			- <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍</span>  <br>String 是最基本的 key-value 结构,key 是唯一标识,value 是具体的值,value其实不仅是字符串, 也可以是数字(整数或浮点数), value 最多可以容纳的数据⻓度是 5 1 2 M [](marginnote3app://note/CEB48A13-3D81-4E8F-998E-D8E2A748029C)
			 - **字符串对象的内部编码(encoding)有 3 种 :int、raw和 embstr。**[](marginnote3app://note/31700CA8-DEBF-4C8C-BDF8-BC317DB386F9)
			 - 整数值,并且这个整数值可以用l o n g 类型来表示,那么字符串对象会将整数值保存在字符串对象结构的p t r 属性里面(将v o i d * 转换成 long),并将字符串对象的编码设置为i n t 。[](marginnote3app://note/AD95B4B2-CC45-438A-8741-B5F9E6C8AD12)
			 - 如果字符串对象保存的是一个字符串,并且这个字符申的⻓度小于等于 32 字节(redis 2.+版本),那么字符串对象将使用一个简单动态字符串(SDS)来保存这个字符串,并将对象的编码设置为**embstr**[](marginnote3app://note/F18EF1F8-CA13-415D-8105-8B026211DEAC)<span><div style="
	border: solid;border: 3px solid;border-image: linear-gradient(to right, #9932CC, #00FFFF)1;border-width:fit-content;text-align:left;padding:0px 0px;"><ul><li>字符串,并且这个字符串的⻓度大于32 字节(redis 2.+版本),那么字符串对象将使用一个简单动态字符串(SDS)来保存这个字符串,并将对象的编码设置为r a w </li><li>redis 2.+ 是 32 字节redis 3.0-4.0 是 39 字节redis 5.0 是 44 字节【源码文件object.c 120】</li><li> e m b s t r 和r a w 编码都会使用S D S 来保存值,但不同之处在于e m b s t r 会通过一次内存分配函数来分配一块连续的内存空间来保存r e d i s O b j e c t 和S D S ,而r a w 编码会通过调用两次内存分配函数来分别分配两块空间来保存r e d i s O b j e c t 和S D S 。</li></ul></div></span>
			 - <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内部实现<br></span>String 类型的底层的数据结构实现主要是 int 和 SDS(简单动态字符串)。[](marginnote3app://note/36DEC904-05E1-43F1-9312-5211DC1D834F)
			
			 -   SDS 相比于 C 的原生字符串:[](marginnote3app://note/811C8849-5D2C-4026-BFDF-F6130BD2FD05)<span><div style="
	border: solid;border: 3px solid;border-image: linear-gradient(to right, #9932CC, #00FFFF)1;border-width:fit-content;text-align:justify;padding:5px 20px;">
			 - SDS 不仅可以保存文本数据,还可以保存二进制数据。<br>
			 - SDS 获取字符串⻓度的时间复杂度是 O(1)。<br>     
			 - Redis 的 SDS API 是安全的,拼接字符串不会造成缓冲区溢出。</div></span>
			 - <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景</span>
			 -  缓存对象[](marginnote3app://note/B560A5B7-7BE5-4490-AEC5-24F1A2701BF4)
			 -  常规计数[](marginnote3app://note/8A6347DE-D93E-4613-87CC-55BE655D3935)<span><div style="
	border: solid;border: 3px solid;border-image: linear-gradient(to right, #9932CC, #00FFFF)1;border-width:fit-content;text-align:left;padding:5px 20px;">因为 Redis 处理命令是单线程,所以执行命令的过程是原子的。因此String 数据类型适合计数场景,比如计算访问次数、点赞、转发、库存数量等等</div></span>
			 -  分布式锁[](marginnote3app://note/F0D539D4-A84B-437D-A582-327898117E25)<span><div style="
	border: solid;border: 3px solid;border-image: linear-gradient(to right, #9932CC, #00FFFF)1;border-width:fit-content;text-align:left;padding:5px 20px;">SET 命令有个 NX 参数可以实现「key不存在才插入」,可以用它来实现分布式锁: <br>如果 key 不存在,则显示插入成功,可以用来表示加锁成功; <br>如果 key 存在,则会显示插入失败,可以用来表示加锁失败。</div></span>
			 - 共享 Session 信息[](marginnote3app://note/8EC92DB8-9577-4D2E-9661-4ED4D9F5DFCD)<span><div style="
	border: solid;border: 3px solid;border-image: linear-gradient(to right, #9932CC, #00FFFF)1;border-width:fit-content;text-align:left;padding:5px 20px;"> 借助 Redis 对这些 Session 信息进行统一的存储和管理,这样无论请求发送到那台服务器,服务器都会去同一个 Redis 获取相关的 Session 信息,这样就解决了分布式系统下 Session 存储的问题。</div></span>
		-  <span class = "t1">List</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍<br></span>
			- List 列表是简单的字符串列表,按照插入顺序排序,可以从头部或尾部向 List 列表添加元素。 列表的最大⻓度为$2^{32}-1$,也即每个列表支持超过 4 0  亿个元素。<span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内部实现<br></span><span class = "bq">List 类型的底层数据结构是由双向链表或压缩列表实现的; 但是在 Redis 3.2 版本之后,List 数据类型底层数据结构就只由<mark>quicklist</mark> 实现了,替代了双向链表和压缩列表。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景<br></span>
			- #消息队列
			- 消息队列在存取消息时,必须要满足三个需求,分别是<b class="focus1">消息保序、处理重复的消息和保证消息可靠性。</b><span><div class="border2"><b> 消息保序</b>List 可以使用 LPUSH + RPOP (或者反过来,RPUSH+LPOP)命令实现消息队列。
			生产者使用 LPUSH  key  value [ value . . . ]  将消息插入到队列的头部,如果 key 不存在则会创建一个空的队列再插入消息。
			消费者使用 RPOP  key  依次读取队列的消息,先进先出。<br> <b>性能⻛险点</b>在生产者往 List 中写入数据时,List 并不会主动地通知消费者有新消息写入,如果消费者想要及时处理消息,就需要在程序中不停地调用R P O P  命令(比如使用一个while(1)循环)。
			解决这个问题,Redis提供了 BRPOP 命令。BRPOP命令也称为阻塞式读取,客户端在没有读到队列数据时,自动阻塞,直到有新的数据写入队列,再开始读取新数据。<br> <b>处理重复的消息</b>每个消息都有一个全局的 ID。;
			 消费者要记录已经处理过的消息的 ID。
			 List 并不会为每个消息生成 ID 号,所以我们需要自行为每个消息生成一个全局唯一ID,生成之后,我们在用 LPUSH 命令把消息插入 List 时,需要在消息中包含这个全局唯一 ID。<br> <b>消息可靠性</b>List 类型提供了 BRPOPLPUSH  命令,这个命令的作用是让消费者程序从一个 List 中读取消息,同时,Redis 会把这个消息再插入到另一个 List(可以叫作备份 List)留存。</div></span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">List 作为消息队列有什么缺陷?<br></span>
			- **List 不支持多个消费者消费同一条消息**,因为一旦消费者拉取一条消息后,这条消息就从 List 中删除了,无法被其它消费者再次消费。		 要实现一条消息可以被多个消费者消费,那么就要将多个消费者组成一个消费组,使得多个消费者可以消费同一条消息,但是 **List 类型并不支持消费组的实现**。
			- 这就要说起 Redis 从 5.0 版本开始提供的 Stream 数据类型了, Stream 同样能够满足消息队列的三大需求,而且它还支持「消费组」形式的消息读取。<span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">Hash</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍<br></span><span class = "bq">Hash 是一个键值对(key - value)集合,其中 value 的形式如: value = [ { field1 ,value1 } ,. . . { fieldN ,valueN } ] 。Hash 特别适合用于存储对象。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内部实现<br></span><span class = "bq">Hash 类型的底层数据结构是由压缩列表或哈希表实现的: 如果哈希类型元素个数小于 5 1 2  个(默认值,可由 hash - maxziplist - entries  配置),所有值小于 6 4  字节(默认值,可由hash - max - ziplist - value  配置)的话,Redis 会使用压缩列表作为 Hash 类型的底层数据结构; 如果哈希类型元素不满足上面条件,Redis 会使用哈希表作为Hash 类型的 底层数据结构。
			在 Redis 7.0 中,压缩列表数据结构已经废弃了,交由 <mark>listpack</mark> 数据结构来实现了。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景<br></span>
			- **缓存对象**; Hash 类型的 (key,field, value) 的结构与对象的(对象id, 属性, 值)的结构相似,也可以用来存储对象。;  
			- **购物⻋**; 以用户 id 为 key,商品 id 为 field,商品数量为 value,恰好构成了购物⻋的3个要素; 涉及的命令如下: 
			 <code class="sum" >添加商品:
HSET cart:{用户id}{商品id}1
添加数量:HINCRBYcart:{用户id}{商品id}1
商品总数:HLENcart:{用户id}
删除商品:HDELcart:{用户id}{商品id}
获取购物⻋所有商品:HGETALLcart:{用户id}
当前仅仅是将商品ID存储到了Redis 中,在回显商品具体信息的时候,还需要拿着商品 id 查询一次数据库,获取完整的商品的信息。</code><span style="display:block;margin-bottom:2%;"></span>
		  -  <span class = "t1">Set</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍<br></span>Set 类型是一个无序并唯一的键值集合,它的存储顺序不会按照插入的先后顺序进行存储。; 一个集合最多可以存储 $2 ^ {3 2} - 1$  个元素。<span><table ><caption>Set 类型和 List 类型的区别如下</caption><tr><td>List 可以存储重复元素</td><td>Set 只能存储非重复元素;</td><tr> <td>List 是按照元素的先后顺序存储元素的,</td><td>而 Set 则是无序方式存储元素的。</td></tr></table></span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内部实现<br></span><span class = "bq">Set 类型的底层数据结构是由哈希表或整数集合实现的: 如果集合中的元素都是整数且元素个数小于 5 1 2  (默认值,s e t m a x i n t s e t - e n t r i e s 配置)个,Redis 会使用整数集合作为 Set 类型的底层数据结构; 如果集合中的元素不满足上面条件,则 Redis 使用哈希表作为Set 类型的底层数据结构。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景<br></span>集合的主要几个特性,无序、不可重复、支持并交差等操作。<b class="genre1"> Set 的差集、并集和交集的计算复杂度较高,在数据量较大的情况下,如果直接执行这些计算,会导致 Redis 实例阻塞;  </b><br>**点赞**; Set 类型可以保证一个用户只能点一个赞,这里举例子一个场景,key 是文章id,value 是用户id。<br> **共同关注**; Set 类型支持交集运算,所以可以用来计算共同关注的好友、公众号等。<br> **抽奖活动**; 存储某活动中中奖的用户名 ,Set 类型因为有去重功能,可以保证同一个用户不会中奖两次。
   如果允许重复中奖,可以使用 SRANDMEMBER 命令。
   如果不允许重复中奖,可以使用 SPOP 命令。<span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">Zset</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍<br></span><span class = "bq">Zset 类型(有序集合类型)相比于 Set 类型多了一个排序属性 score (分值),对于有序集合 ZSet 来说,每个存储元素相当于有两个值组成的,一个是有序集合的元素值,一个是排序值。
		 有序集合保留了集合不能有重复成员的特性(分值可以重复),但不同的是,有序集合中的元素可以排序。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内部实现<br></span><span class = "bq">Zset 类型的底层数据结构是由压缩列表或跳表实现的: 如果有序集合的元素个数小于 1 2 8  个,并且每个元素的值小于6 4  字节时,Redis 会使用压缩列表作为 Zset 类型的底层数据结构; 如果有序集合的元素不满足上面的条件,Redis 会使用跳表作为Zset 类型的底层数据结构;; 在 Redis 7.0 中,压缩列表数据结构已经废弃了,交由 <em>listpack</em> 数据结构来实现了。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景<br></span><b class="genre1">Zset 类型(Sorted Set,有序集合) 可以根据元素的权重来排序,我们可以自己来决定每个元素的权重值。</b>; 
			- **排行榜**; 有序集合比较典型的使用场景就是排行榜。; 
			- **电话、姓名排序**; 使用有序集合的 Z R A N G E B Y L E X  或 Z R E V R A N G E B Y L E X  可以帮助我们实现电话号码或姓名的排序,我们以 Z R A N G E B Y L E X  (返回指定成员区间内的成员,按 key 正序排列,分数必须相同)为例。<br>	 注意:不要在分数不一致的 SortSet 集合中去使用 ZRANGEBYLEX 和 ZREVRANGEBYLEX 指令,因为获取的结果会不准确。<span style="display:block;margin-bottom:2%;"></span>
		-  **BitMap(2.2 版新增)、HyperLogLog(2.8 版新增)、GEO(3.2 版新增)、Stream(5.0 版新增)
		 -  <span class = "bq">BitMap</span> **<span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍<br></span><span class = "bq">Bitmap,即位图,是一串连续的二进制数组(0和1),可以通过偏移量(offset)定位元素。BitMap通过最小的单位bit来进行0 | 1 的设置, 表示某个元素的值或者状态,时间复杂度为O(1); 由于 bit 是计算机中最小的单位,使用它进行储存将非常节省空间, 特别适合一些数据量大且使用二值统计的场景。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内部实现<br></span><span class = "bq">Bitmap 本身是用 String 类型作为底层数据结构实现的一种统计二值状态的数据类型。
		 String 类型是会保存为二进制的字节数组,所以,Redis 就把字节数组的每个 bit 位利用起来,用来表示一个元素的二值状态,你可以把Bitmap 看作是一个 bit 数组。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景<br></span>
			- <b class="genre1">Bitmap 类型非常适合二值状态统计的场景,这里的二值状态就是指集合元素的取值就只有 0 和 1 两种,在记录海量数据时,Bitmap 能够有效地节省内存空间。</b><br>
			- **签到统计** 在签到打卡的场景中,我们只用记录签到(1)或未签到(0),所以它就是非常典型的二值状态。<span><div class="border1"> 如何统计这个月首次打卡时间呢?<br>Redis 提供了 <code>BITPOSkeybitValue[start][end]</code> 指令,返回数据表示 Bitmap 中第一个值为 b i t V a l u e  的 offset 位置。 在默认情况下, 命令将检测整个位图, 用户可以通过可选的 s t a r t 参数和 e n d  参数指定要检测的范围。; 需要注意的是,因为 offset 从 0 开始的,所以我们需要将返回的value + 1 。</div></span>
			- **判断用户登陆态**; Bitmap 提供了 G E T B I T 、S E T B I T  操作,通过一个偏移值 offset 对 bit 数组的 offset 位置的 bit 位进行读写操作,需要注意的是 offset 从 0 开始。; 只需要一个 key = login_status 表示存储用户登陆状态集合数据, 将用户 ID 作为 offset,在线就设置为 1,下线设置 0。通过 G E T B I T 判断对应的用户是否在线。 ==5000 万用户只需要 6 MB 的空间。==
			- **连续签到用户总数** 我们把每天的日期作为 Bitmap 的 key,userId 作为 offset,若是打卡则将 offset 位置的 bit 设置成 1。<span><div class="border1"> key 对应的集合的每个 bit 位的数据则是一个用户在该日期的打卡记录。<br> 一共有 7 个这样的 Bitmap,如果我们能对这 7 个 Bitmap 的对应的bit 位做 『与』运算。同样的 UserID offset 都是一样的,当一个userID 在 7 个 Bitmap 对应对应的 offset 位置的 bit = 1 就说明该用户7 天连续打卡。<br> 结果保存到一个新 Bitmap 中,我们再通过 B I T C O U N T  统计 bit = 1 的个数便得到了连续打卡 7 天的用户总数了。<br>		 Redis 提供了 BITOPoperationdestkeykey[key...] 这个指令用于对一个或者多个 key 的 Bitmap 进行位元操作。<br>		 o p e r a t i o n  可以是 a n d 、O R 、N O T 、X O R 。当 BITOP 处理不同⻓度的字符串时,较短的那个字符串所缺少的部分会被看作 0  。空的k e y  也被看作是包含 0  的字符串序列。<br>		 假设要统计 3 天连续打卡的用户数,则是将三个 bitmap 进行 AND 操作,并将结果保存到 destmap 中,接着对 destmap 执行 BITCOUNT 统计,</div></span><span style="display:block;margin-bottom:2%;"></span>
		 -  <span class = "t1">HyperLogLog</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍<br></span>
			- Redis HyperLogLog 是 Redis 2.8.9 版本新增的数据类型,是一种用于「统计基数」的数据集合类型,基数统计就是指统计一个集合中不重复的元素个数。但要注意,HyperLogLog 是统计规则是基于概率完成的,不是非常准确,标准误算率是 0.81%。; 
			- **HyperLogLog 的优点是**,在输入元素的数量或者体积非常非常大时, 计算基数所需的内存空间总是固定的、并且是很小的; ==在 Redis 里面,每个 HyperLogLog 键只需要花费 12 KB 内存,就可以计算接近 $2^{64}$ 个不同元素的基数==,和元素越多就越耗费内存的 Set 和 Hash 类型相比,HyperLogLog 就非常节省空间。; 简单来说 HyperLogLog 提供不精确的去重计数。 <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内部实现<br></span><span class = "bq">HyperLogLog 的实现涉及到很多数学问题,太费脑子了,我也没有搞懂,如果你想了解一下,课下可以看看这个</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景<br></span><span class = "bq">百万级网⻚ UV 计数</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">GEO</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍<br></span><span class = "bq">Redis GEO 是 Redis 3.2 版本新增的数据类型,主要用于存储地理位置信息,并对存储的信息进行操作。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内部实现<br></span>
			- GEO 本身并没有设计新的底层数据结构,而是直接使用了 Sorted Set 集合类型。; GEO 类型使用 GeoHash 编码方法实现了经纬度到 Sorted Set 中元素权重分数的转换,这其中的两个关键机制就是「对二维地图做区间划分」和「对区间进行编码」。一组经纬度落在某个区间后,就用区间的编码值来表示,并把编码值作为 Sorted Set 元素的权重分数。	
			- 这样一来,我们就可以把经纬度保存到 Sorted Set 中,利用 Sorted Set 提供的“按权重进行有序范围查找”的特性,实现 LBS 服务中频繁使用的“搜索附近”的需求。<span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景<br></span><span class = "bq"><strong>滴滴叫⻋</strong>这里以滴滴叫⻋的场景为例,介绍下具体如何使用 GEO 命令: GEOADD 和 GEORADIUS 这两个命令。假设⻋辆 ID 是 33,经纬度位置是(116.034579,39.030452),我们可以用一个 GEO 集合保存所有⻋辆的经纬度,集合 key 是cars:locations。
			 当用户想要寻找自己附近的网约⻋时,LBS 应用就可以使用GEORADIUS 命令。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">Stream</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">介绍<br></span><span class = "bq">Redis Stream 是 Redis 5.0 版本新增加的数据类型,Redis 专⻔为消息队列设计的数据类型。
		 在 Redis 5.0 Stream 没出来之前,消息队列的实现方式都有着各自的缺陷,例如: 发布订阅模式,不能持久化也就无法可靠的保存消息,并且对于离线重连的客户端不能读取历史消息的缺陷; List 实现消息队列的方式不能重复消费,一个消息消费完就会被删除,而且生产者需要自行实现全局唯一 ID。
		 基于以上问题,Redis 5.0 便推出了 Stream 类型也是此版本最重要的功能,用于完美地实现消息队列,它支持消息的持久化、支持自动生成全局唯一 ID、支持 ack 确认消息的模式、支持消费组模式等,让消息队列更加的稳定和可靠。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">应用场景<br></span>
			- #消息队列 <span class = "bq"> 生产者通过 XADD 命令插入一条消息:; 插入成功后会返回全局唯一的 ID:"1654254953808-0"。消息的全局唯一 ID 由两部分组成: <span><div class="border2">第一部分“1654254953808”是数据插入时,以毫秒为单位计算的当前服务器时间; 第二部分表示插入消息在当前毫秒内的消息序号,这是从 0 开始编号的。例如,“1654254953808-0”就表示在“1654254953808”毫秒内的第 1 条消息。
		 消费者通过 XREAD 命令从消息队列中读取消息时,可以指定一个消息 ID,并从这个消息 ID 的下一条消息开始进行读取(注意是输入消息 ID 的下一条信息开始读取,不是查询输入ID的消息)。</div></span>
		  如果想要实现阻塞读(当没有数据时,阻塞住),可以调用 XRAED 时设定 BLOCK 配置项,实现类似于 BRPOP 的阻塞读取操作。
前面介绍的这些操作 List 也支持的,接下来看看 Stream 特有的功能。
Stream 可以以使用 XGROUP 创建消费组,创建消费组之后,Stream 可以使用 XREADGROUP 命令让消费组内的消费者读取消息。
<b class="genre1">消息队列中的消息一旦被消费组里的一个消费者读取了,就不能再被该消费组内的其他消费者读取了,即同一个消费组里的消费者不能消费同一条消息。
但是,不同消费组的消费者可以消费同一条消息(但是有前提条件, 创建消息组的时候,不同消费组指定了相同位置开始读取消息)。</b>
使用消费组的目的是让组内的多个消费者共同分担读取消息,所以, 我们通常会让每个消费者读取部分消息,从而实现消息读取负载在多个消费者间是均衡分布的。
<strong>基于 Stream 实现的消息队列,如何保证消费者在发生故障或宕机再次重启后,仍然可以读取未处理完的消息?;</strong>
Streams 会自动使用内部队列(也称为 PENDING List)留存消费组里每个消费者读取的消息,直到消费者使用 XACK 命令通知Streams“消息已经处理完成”。
		 消费确认增加了消息的可靠性,一般在业务处理完成之后,需要执行XACK 命令确认消息已经被消费完成,整个流程的执行如下图所示:
		 
		  ![[Pasted image 20230710192010.png]]
		  
		 如果消费者没有成功处理消息,它就不会给 Streams 发送 XACK 命令,消息仍然会留存。此时,消费者可以在重启后,用 XPENDING 命令查看已读取、但尚未确认处理完成的消息。; 一旦消息 1654256265584-0 被 consumer2 处理了,consumer2 就可以使用 XACK 命令通知 Streams,然后这条消息就会被删除。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 基于 Stream 消息队列与专业的消息队列有哪些差距?<br></span><span class = "bq">Redis 基于 Stream 消息队列与专业的消息队列有哪些差距?; 一个专业的消息队列,必须要做到两大块: 消息不丢。
		 消息可堆积。
		  <ol start="1"><li>Redis Stream 消息会丢失吗? </li>
		  使用一个消息队列,其实就分为三大块:生产者、队列中间件、消费者,所以要保证消息就是保证三个环节都不能丢失数据。
		  <span><div class="border1"> 消息生产阶段消息存储阶段send消息生产者ack消息消费阶段pull消息中间件消息生产者ack; Redis Stream 消息队列能不能保证三个环节都不丢失数据? <ul>
		  <li>Redis 生产者会不会丢消息?生产者会不会丢消息,取决于生产者对于异常情况的处理是否合理。 从消息被生产出来,然后提交给 MQ 的过程中,只要能正常收到 ( MQ 中间件) 的 ack 确认响应,就表示发送成功,所以只要处理好返回值和异常,如果返回异常则进行消息重发,那么这个阶段是不会出现消息丢失的。</li>
		<li> Redis 消费者会不会丢消息?不会,因为 Stream ( MQ 中间件)会自动使用内部队列(也称为 PENDING List)留存消费组里每个消费者读取的消息,但是未被确认的消息。消费者可以在重启后,用 XPENDING 命令查看已读取、但尚未确认处理完成的消息。等到消费者执行完业务逻辑后,再发送消费确认 XACK 命令,也能保证消息的不丢失。</li>
		<li> Redis 消息中间件会不会丢消息?会,Redis 在以下 2 个场景下,都会导致数据丢失: <ul><li>AOF 持久化配置为每秒写盘,但这个写盘过程是异步的, Redis 宕机时会存在数据丢失的可能</li><li>主从复制也是异步的,主从切换时,也存在丢失数据的可能。</li></ul></li></ul>
		 可以看到,Redis 在队列中间件环节无法保证消息不丢。像RabbitMQ 或 Kafka 这类专业的队列中间件,在使用时是部署一个集群,生产者在发布消息时,队列中间件通常会写「多个节点」,也就是有多个副本,这样一来,即便其中一个节点挂了,也能保证集群的数据不丢失。</div></span>
		<li> Redis Stream 消息可堆积吗? </li>Redis 的数据都存储在内存中,这就意味着一旦发生消息积压,则会导致 Redis 的内存持续增⻓,如果超过机器内存上限,就会面临被OOM 的⻛险。
		 所以 Redis 的 Stream 提供了可以指定队列最大⻓度的功能,就是为了避免这种情况发生。
		 当指定队列最大⻓度时,队列⻓度超过上限后,旧消息会被删除,只保留固定⻓度的新消息。这么来看,Stream 在消息积压时,如果指定了最大⻓度,还是有可能丢失消息的。
		 但 Kafka、RabbitMQ 专业的消息队列它们的数据都是存储在磁盘上,当消息积压时,无非就是多占用一些磁盘空间。
		 因此,把 Redis 当作队列来使用时,会面临的 2 个问题: <ul><li>Redis 本身可能会丢数据;</li><li> 面对消息挤压,内存资源会紧张;</li></ul>
		  所以,能不能将 Redis 作为消息队列来使用,关键看你的业务场景:<ul><li> 如果你的业务场景足够简单,对于数据丢失不敏感,而且消息积压概率比较小的情况下,把 Redis 当作队列是完全可以的。</li>
		 <li>如果你的业务有海量消息,消息积压的概率比较大,并且不能接受数据丢失,那么还是用专业的消息队列中间件吧。</li></ul>
		 补充:Redis 发布/订阅机制为什么不可以作为消息队列? 发布订阅机制存在以下缺点,都是跟丢失数据有关: 
		 1. 发布/订阅机制没有基于任何数据类型实现,所以不具备「数据持久化」的能力,也就是发布/订阅机制的相关操作,不会写入到 RDB 和 AOF 中,当 Redis 宕机重启,发布/订阅机制的数据也会全部丢失。
		 2. 发布订阅模式是“发后既忘”的工作模式,如果有订阅者离线重连之后不能消费之前的历史消息。
		 3. 当消费端有一定的消息积压时,也就是生产者发送的消息,消费者消费不过来时,如果超过 32M 或者是 60s 内持续保持在 8M 以上,消费端会被强行断开,这个参数是在配置文件中设置的, 默认值是client-output-buffer-limit pubsub 32mb 8mb 60。
		 所以,发布/订阅机制只适合即时通讯的场景,比如构建哨兵集群 的场景采用了发布/订阅机制。</ol></span> <span style="display:block;margin-bottom:2%;"></span>

 -  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">之后源码中查看原因，加深印象<br></span> ^bce612
	 - <span class = "bq">embstr 编码将创建字符串对象所需的内存分配次数从 raw  编码的两次降低为一次; 释放 embstr 编码的字符串对象同样只需要调用一次内存释放函数; 因为embstr 编码的字符串对象的所有数据都保存在一块连续的内存里面可以更好的利用 CPU 缓存提升性能。
	 但是 embstr 也有缺点的: 如果字符串的⻓度增加需要重新分配内存时,整个redisObject和sds都需要重新分配空间,所以embstr编码的字符串对象实际上是只读的,redis没有为embstr编码的字符串对象编写任何相应的修改程序。当我们对embstr编码的字符串对象执行任何修改命令(例如append)时,程序会先将对象的编码从embstr转换成raw,然后再执行修改命令</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">一般而言,还会对分布式锁加上过期时间,分布式锁的命令如下:`SET lock_key  unique_value NX PX  10000` 
		lock_key 就是 key 键;
		unique_value 是客户端生成的唯一的标识;
		NX 代表只在 lock_key 不存在时,才对 lock_key 进行设置操作;
		PX 10000 表示设置 lock_key 的过期时间为 10s,这是为了避免客户端发生异常而无法释放锁。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">基于 List 类型的消息队列,满足消息队列的三大需求(消息保序、处理重复的消息和保证消息可靠性)。
		 消息保序:使用 LPUSH + RPOP; 阻塞读取:使用 BRPOP; 重复消息处理:生产者自行实现全局唯一 ID; 消息的可靠性:使用 BRPOPLPUSH</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">基于 Stream 实现的消息队列就说到这里了,小结一下: 消息保序:XADD/XREAD 阻塞读取:XREAD block 重复消息处理:Stream 在使用 XADD 命令,会自动生成全局唯一 ID; 消息可靠性:内部使用 PENDING List 自动保存消息,使用XPENDING 命令查看消费组已经读取但是未被确认的消息,消费者使用 XACK 确认消息; 支持消费组形式消费数据</span> <span style="display:block;margin-bottom:2%;"></span>
## Redis 数据结构 | 小林coding

-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 数据结构 | 小林coding<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "t1">新旧版本的数据结构说图解一遍,共有 9 种数据结构:SDS、双向链表、压缩列表、哈希表、跳表、整数集合、quicklist、listpack。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "high2">Redis 对象</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">哈希桶存放的是指向键值对数据的指针(dictEntry*),这样通过指针就能找到键值对数据,然后因为键值对的值可以保存字符串对象和集合数据类型的对象, 所以键值对的数据结构中并不是直接保存值本身,而是保存了 void * key 和 void * value 指针,分别指向了实际的键对象和值对象,这样一来,即使值是集合数据,也可以通过 void * value 指针找到。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 中的每个对象都由 redisObject 结构<br></span><span class = "bq">对象结构里包含的成员变量: type,标识该对象是什么类型的对象(String 对象、 List 对象、Hash 对象、Set 对象和 Zset 对象); encoding,标识该对象使用了哪种底层的数据结构; ptr,指向底层数据结构的指针。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">SDS</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">C 语言的字符串不足之处以及可以改进的地方:<br></span><span class = "bq">获取字符串⻓度的时间复杂度为 O(N); 字符串的结尾是以 “\0” 字符标识,字符串里面不能包含有 “\0” 字符,因此不能保存二进制数据; 字符串操作函数不高效且不安全,比如有缓冲区溢出的⻛险,有可能会造成程序运行终止;</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">SDS 结构设计<br></span><span class = "bq">len,记录了字符串⻓度。这样获取字符串⻓度的时候,只需要返回这个成员变量值就行,时间复杂度只需要 O(1)。
 alloc,分配给字符数组的空间⻓度。这样在修改字符串的时候,可以通过alloc - len  计算出剩余的空间大小,可以用来判断空间是否满足修改需求,<strong>如果不满足的话,就会自动将 SDS 的空间扩展至执行修改所需的大小</strong>, 然后才执行实际的修改操作,所以使用 SDS 既不需要手动修改 SDS 的空间大小,也不会出现前面所说的缓冲区溢出的问题。
 flags,用来表示不同类型的 SDS。一共设计了 5 种类型,分别是 sdshdr5、sdshdr8、sdshdr16、sdshdr32 和 sdshdr64,后面在说明区别之处。
 buf[],字符数组,用来保存实际数据。不仅可以保存字符串,也可以保存二进制数据。</span> <span style="display:block;margin-bottom:2%;"></span><span><div class="border1">
				-  <span class = "bq">O(1)复杂度获取字符串⻓度</span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span class = "bq">二进制安全</span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">不会发生缓冲区溢出<br></span><span class = "bq">Redis 的 SDS 结构里引入了 alloc 和 len 成员变量,这样 SDS API 通过alloc - len  计算,可以算出剩余可用的空间大小,这样在对字符串做修改操作的时候,就可以由程序内部判断缓冲区大小是否足够用。</span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span class = "bq">__attribute__ ((packed))  ,它的作用是:告诉编译器取消结构体在编译过程中的优化对⻬,按照实际占用字节数进行对⻬。</span></div></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">链表</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">链表节点结构设计<br></span><span class = "bq">这个是一个双向链表。</span> <span style="display:block;margin-bottom:2%;"></span>
```cpp
   typedef struct listNode {
//前置节点
    struct listNode *prev;
//后置节点
    struct listNode *next;
//节点的值
    void *value;

} listNode; 
```

<br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">链表结构设计<br></span><span class = "bq">list 结构为链表提供了链表头指针 head、链表尾节点 tail、链表节点数量 len、以及可以自定义实现的 dup、free、match 函数。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">链表的优势与缺陷<br></span><span class = "bq">Redis 的链表实现<br><div class="border2"><strong>优点</strong>如下: <ul><li>listNode 链表节点的结构里带有 prev 和 next 指针,获取某个节点的前置节点或后置节点的时间复杂度只需O(1),而且这两个指针都可以指向 NULL,所以链表是无环链表;</li><li>list 结构因为提供了表头指针 head 和表尾节点 tail,所以获取链表的表头节点和表尾节点的时间复杂度只需O(1);</li><li>list 结构因为提供了链表节点数量 len,所以获取链表中的节点数量的时间复杂度只需O(1);</li><li>listNode 链表节使用 void* 指针保存节点值,并且可以通过 list 结构的 dup、free、match 函数指针为节点设置该节点类型特定的函数,因此链表节点可以保存各种不同类型的值; </li></ul>链表的<strong>缺陷</strong>也是有的: <ul><li>链表每个节点之间的内存都是不连续的,意味着无法很好利用 CPU 缓存。能很好利用 CPU 缓存的数据结构就是数组,因为数组的内存是连续的,这样就可以充分利用 CPU 缓存来加速访问。</li><li>还有一点,保存一个链表节点的值都需要一个链表节点结构头的分配,内存开销较大。</li></ul></div></span><span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">压缩列表</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">压缩列表的最大特点,就是它被设计成一种内存紧凑型的数据结构,占用一块连续的内存空间,不仅可以利用 CPU 缓存,而且会针对不同⻓度的数据,进行相应编码,这种方法可以有效地节省内存开销。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">压缩列表的缺陷也是有的: 不能保存过多的元素,否则查询效率就会降低; 新增或修改某个元素时,压缩列表占用的内存空间需要重新分配,甚至可能引发连锁更新的问题。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">压缩列表结构设计<br></span><span class = "bq">压缩列表在表头有三个字段: zlbytes,记录整个压缩列表占用对内存字节数; zltail,记录压缩列表「尾部」节点距离起始地址由多少字节,也就是列表尾的偏移量; zllen,记录压缩列表包含的节点数量; zlend,标记压缩列表的结束点,固定值 0xFF(十进制255)。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">压缩列表节点包含三部分内容:<br></span><span class = "bq">prevlen,记录了「前一个节点」的⻓度,目的是为了实现从后向前遍历; encoding,记录了当前节点实际数据的「类型和⻓度」,类型主要有两种: 字符串和整数。
 data,记录了当前节点的实际数据,类型和⻓度都由 encoding  决定;</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">连锁更新<br></span><span class = "bq">压缩列表新增某个元素或修改某个元素时,如果空间不不够,压缩列表占用的内存空间就需要重新分配。而当新插入的元素较大时,可能会导致后续元素的prevlen 占用空间都发生变化,从而引起「连锁更新」问题,导致每个元素的空间都要重新分配,造成访问压缩列表性能的下降。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">这种在特殊情况下产生的连续多次空间扩展操作就叫做「连锁更新」,</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "high2">哈希表</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">Redis 采用了「链式哈希」来解决哈希冲突,在不扩容哈希表的前提下,将具有相同哈希值的数据串起来,形成链接起,以便这些数据在表中仍然可以被查询到。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">rehash<br></span><span class = "bq">给「哈希表 2」 分配空间,一般会比「哈希表 1」 大 2 倍;
			 将「哈希表 1 」的数据迁移到「哈希表 2」 中; 迁移完成后,「哈希表 1 」的空间会被释放,并把「哈希表 2」 设置为「哈希表 1」,然后在「哈希表 2」 新创建一个空白的哈希表,为下次 rehash 做准备。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">渐进式 rehash<br></span><span class = "bq">给「哈希表 2」 分配空间; 在 rehash 进行期间,每次哈希表元素进行新增、删除、查找或者更新操作时,Redis 除了会执行对应的操作之外,还会顺序将「哈希表 1 」中索引位置上的所有 key-value 迁移到「哈希表 2」 上;; 随着处理客户端发起的哈希表操作请求数量越多,最终在某个时间点会把「哈希表 1 」的所有 key-value 迁移到「哈希表 2」,从而完成 rehash 操作。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">触发 rehash 操作的条件<br></span><span class = "bq">当负载因子大于等于 1 ,并且 Redis 没有在执行 bgsave 命令或者bgrewiteaof 命令,也就是没有执行 RDB 快照或没有进行 AOF 重写的时候, 就会进行 rehash 操作。
 当负载因子大于等于 5 时,此时说明哈希冲突非常严重了,不管有没有有在执行 RDB 快照或 AOF 重写,都会强制进行 rehash 操作。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">整数集合</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">整数集合是 Set 对象的底层实现之一。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">整数集合的升级操作<br></span><span class = "bq">整数集合会有一个升级规则,就是当我们将一个新元素加入到整数集合里面,如果新元素的类型(int32_t)比整数集合现有所有元素的类型(int16_t)都要⻓时,整数集合需要先进行升级,也就是按新元素的类型(int32_t)扩展contents 数组的空间大小,然后才能将新元素加入到整数集合里,当然升级的过程中,也要维持整数集合的有序性。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "high2">跳表</span> <span style="display:block;margin-bottom:2%;"></span>
			-  Redis 只有 Zset 对象的底层实现用到了跳表,跳表的优势是能支持平均 $O(log^N)$ 复杂度的节点查找。<span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">跳表结构<br></span><span class = "bq">跳表的头尾节点,便于在O(1)时间复杂度内访问跳表的头节点和尾节点; 跳表的⻓度,便于在O(1)时间复杂度获取跳表节点的数量; 跳表的最大层数,便于在O(1)时间复杂度获取跳表中层高最大的那个节点的层数量;</span> <span style="display:block;margin-bottom:2%;"></span>
			-  跳表的相邻两层的节点数量最理想的比例是 2:1,查找复杂度可以降低到$O(log^N)$。<span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">跳表在创建节点时候,会生成范围为[0-1]的一个随机数,如果这个随机数小于 0.25(相当于概率 25%),那么层数就增加 1 层,然后继续生成下一个随机数,直到随机数的结果大于 0.25 结束,最终确定该节点的层数。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">quicklist</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">在 Redis 3.0 之前,List 对象的底层数据结构是双向链表或者压缩列表。然后在Redis 3.2 的时候,List 对象的底层改由 quicklist 数据结构实现。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">通过控制每个链表节点中的压缩列表的大小或者元素个数, 来规避连锁更新的问题。因为压缩列表元素越少或越小,连锁更新带来的影响就越小,从而提供了更好的访问性能。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "t1">listpack</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">Redis 在 5.0 新设计一个数据结构叫 listpack,目的是替代压缩列表,它最大特点是 listpack 中每个节点不再包含前一个节点的⻓度了,压缩列表每个节点正因为需要保存前一个节点的⻓度字段,就会有连锁更新的隐患。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">encoding,定义该元素的编码类型,会对不同⻓度的整数和字符串进行编码; data,实际存放的数据; len,encoding+data的总⻓度;</span> <span style="display:block;margin-bottom:2%;"></span>

## 深入理解Redis跳跃表的基本实现和特性 - 掘金    19

-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">深入理解Redis跳跃表的基本实现和特性 - 掘金<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  所以,跳表(skip list)对表的是平衡树(AVL Tree)和 二分查找,是一种 插入/删除/搜索 都是 $O(log^n)$  的数据结构。1989 年出现。<span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">它的最大优势是原理简单、容易实现、方便扩展、效率更高。因此在一些热⻔的项目里用来替代平衡树,如 Redis 、 LevelDB 等。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">跳跃表的构成<br></span><span class = "bq"><div class="border1">表头(head):  负责维护跳跃表的节点指针。
 跳跃表节点:  保存着元素值,以及多个层。
 层:  保存着指向其他元素的指针。高层的指针越过的元素数量大于等于低层的指针,为了提高查找的效率,程序总是从高层先开始访问,然后随着元素值范围的缩小,慢慢降低层次。
 表尾:   全部由 NULL 组成,表示跳跃表的末尾。</div></span> <span style="display:block;margin-bottom:3%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">总结<br></span><span class = "bq">跳跃表是有序集合的底层实现之一, 除此之外它在 Redis 中没有其他应用。
 Redis 的跳跃表实现由 zskiplist  和 zskiplistNode  两个结构组成, 其中 zskiplist  用于保存跳跃表信息(比如表头节点、表尾节点、⻓度), 而 zskiplistNode  则用于表示跳跃表节点。
 每个跳跃表节点的层高都是 1  至 32  之间的随机数。
 在同一个跳跃表中, 多个节点可以包含相同的分值, 但每个节点的成员对象必须是唯一的。</span> <span style="display:block;margin-bottom:2%;"></span>

# 单线程和多线程
## 高性能IO模型：为什么单线程Redis能那么快？.md
-  #IO模型<span style="color:#a85d45;border-bottom:2px solid #a85d45;">03 高性能IO模型：为什么单线程Redis能那么快？.md<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">Redis 是单线程,主要是指 Redis 的网络 IO 和键值对读写是由一个线程来完成的,这也是 Redis 对外提供键值存储服务的主要流程。但 Redis 的其他功能,比如持久化、异步删除、集群数据同步等,其实是由额外的线程执行的。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">深入地学习下 Redis 的单线程设计机制以及多路复用机制。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">多线程的开销<br></span><span class = "bq">“使用多线程,可以增加系统吞吐率,或是可以增加系统扩展性。”的确,对于一个多线程的系统来说,在有合理的资源分配的情况下,可以增加系统中处理请求操作的资源实体,进而提升系统能够同时处理的请求数,即<b class="focus1">吞吐率</b>。
<div class="border2">吞吐率是服务器并发处理能力的量化描述，单位是reqs/s，指的是某个并发用户数下单位时间内处理的请求数。</div>
我们刚开始增加线程数时,系统吞吐率会增加,但是,再进一步增加线程时,系统吞吐率就增⻓迟缓了;
当有多个线程要修改这个共享资源时,为了保证共享资源的正确性,就需要有额外的机制进行保证,而这个额外的机制,就会带来额外的开销。
 而且,采用多线程开发一般会引入同步原语来保护共享资源的并发访问,这也会降低系统代码的易调试性和可维护性。为了避免这些问题,Redis 直接采用了单线程模式。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">单线程 Redis 为什么那么快?<br></span><span class = "bq">一方面,Redis 的大部分操作在内存上完成,再加上它采用了高效的数据结构,例如哈希表和跳表,这是它实现高性能的一个重要原因。另一方面,就是 Redis 采用了多路复用机制,使其在网络IO 操作中能并发处理大量的客户端请求,实现高吞吐率。</span><span style="display:block;margin-bottom:2%;"></span>
	- <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis基本IO模型<br></span> 
	- <span class = "bq">在这里的网络 IO 操作中,有潜在的阻塞点,分别是 accept() 和 recv()。当 Redis 监听到一个客户端有连接请求,但一直未能成功建立起连接时,会阻塞在 accept() 函数这里,导致其他客户端无法和 Redis 建立连接。类似的,当 Redis 通过 recv() 从一个客户端读取数据时,如果数据一直没有到达,Redis 也会一直阻塞在 recv()。</span> <span style="display:block;margin-bottom:2%;"></span><span class = "bq"><div class="border2">以 Get 请求为例,SimpleKV 为了处理一个 Get 请求,需要监听客户端请求(bind/listen),和客户端建立连接(accept),从 socket 中读取请求(recv),解析客户端发送请求(parse),根据请求类型读取键值数据(get),最后给客户端返回结果,即向 socket 中写回数据(send)。
	<b class="focus1">bind/listen、accept、recv、parse 和 send </b>属于网络 IO 处理,而 get 属于键值数据操作。既然 Redis 是单线程,那么,最基本的一种实现是在一个线程中依次执行上面说的这些操作。</div></span>
	
![[Pasted image 20230619151802.png]] 

 <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">非阻塞模式<br></span><span class = "bq">Socket 网络模型的非阻塞模式设置,主要体现在三个关键的函数调用上,如果想要使用 socket 非阻塞模式,就必须要了解这三个函数的调用返回类型和设置模式。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">在 socket 模型中,不同操作调用后会返回不同的套接字类型。socket() 方法会返回主动套接字,然后调用 listen() 方法,将主动套接字转化为监听套接字,此时,可以监听来自客户端的连接请求。
 最后,调用 accept() 方法接收到达的客户端连接,并返回已连接套接字。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">针对监听套接字,我们可以设置非阻塞模式:当 Redis 调用 accept() 但一直未有连接请求到达时, Redis 线程可以返回处理其他操作,而不用一直等待。但是,你要注意的是,调用 accept() 时,已经存在监听套接字了。
 虽然 Redis 线程可以不用继续等待,但是总得有机制继续在监听套接字上等待后续连接请求,并在有请求时通知 Redis。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">类似的,我们也可以针对已连接套接字设置非阻塞模式:Redis 调用 recv() 后,如果已连接套接字上一直没有数据到达,Redis 线程同样可以返回处理其他操作。我们也需要有机制继续监听该已连接套接字,并在有数据达到时通知 Redis。
 这样才能保证 Redis 线程,既不会像基本 IO 模型中一直在阻塞点等待,也不会导致 Redis 无法处理实际到达的连接请求或数据。
 到此,Linux 中的 IO 多路复用机制就要登场了。</span> <span style="display:block;margin-bottom:2%;"></span>
	- <span style="color:#a85d45;border-bottom:2px solid #a85d45;">基于多路复用的高性能 I/O 模型<br></span>
	- <span class = "bq"> <b class="genre1">Linux 中的 IO 多路复用机制是指一个线程处理多个 IO 流,就是我们经常听到的 select/epoll 机制。</b>
 简单来说,在 Redis 只运行单线程的情况下,该机制允许内核中,同时存在多个监听套接字和已连接套接字。内核会一直监听这些套接字上的连接请求或数据请求。一旦有请求到达,就会交给Redis 线程处理,这就实现了一个 Redis 线程处理多个 IO 流的效果。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">下图就是基于多路复用的 Redis IO 模型。图中的多个 FD 就是刚才所说的多个套接字。Redis 网络框架调用 epoll 机制,让内核监听这些套接字。此时,Redis 线程不会阻塞在某一个特定的监听或已连接套接字上,也就是说,不会阻塞在某一个特定的客户端请求处理上。正因为此,Redis 可以同时和多个客户端连接并处理请求,从而提升并发性。</span>

 ![[Pasted image 20230619151825.png]]

<br>
		<span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq"><b class="focus1"> 为了在请求到达时能通知到 Redis 线程,select/epoll 提供了基于事件的回调机制,即针对不同事件的发生,调用相应的处理函数。</b></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">其实,select/epoll 一旦监测到 FD 上有请求到达时,就会触发相应的事件。
 这些事件会被放进一个事件队列,Redis 单程对该事件队列不断进行处理。这样一来,Redis 无需一直轮询是否有请求实际发生,这就可以避免造成 CPU 资源浪费。同时,Redis 在对事件队列中的事件进行处理时,会调用相应的处理函数,这就实现了基于事件的回调。因为 Redis 一直在对事件队列进行处理,所以能及时响应客户端请求,提升 Redis 的响应性能。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">具体解释一下。
 这两个请求分别对应 Accept 事件和 Read 事件,Redis 分别对这两个事件注册 accept 和 get 回调函数。当 Linux 内核监听到有连接请求或读数据请求时,就会触发 Accept 事件和 Read 事件,此时,内核就会回调 Redis 相应的 accept 和 get 函数进行处理。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">需要注意的是,即使你的应用场景中部署了不同的操作系统,多路复用机制也是适用的。因为这个机制的实现有很多种,既有基于 Linux 系统下的 select 和 epoll 实现,也有基于 FreeBSD 的kqueue 实现,以及基于 Solaris 的 evport 实现,这样,你可以根据 Redis 实际运行的操作系统, 选择相应的多路复用实现。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">小结<br></span><span class = "bq">今天,我们重点学习了 Redis 线程的三个问题:“Redis 真的只有单线程吗?”“为什么用单线程?”“单线程为什么这么快?”<b class="genre1">现在,我们知道了,Redis 单线程是指它对网络 IO 和数据读写的操作采用了一个线程,而采用单线程的一个核心原因是避免多线程开发的并发控制问题。单线程的 Redis 也能获得高性能,跟多路复用的 IO 模型密切相关,因为这避免了 accept() 和 send()/recv() 潜在的网络 IO 操作阻塞点。</b></span> <span style="display:block;margin-bottom:2%;"></span>
	
## 面试官：你确定 Redis 是单线程的进程吗？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">面试官：你确定 Redis 是单线程的进程吗？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 是单线程吗?<br></span><b class = "genre1">Redis 单线程指的是「接收客户端请求->解析请求 ->进行数据读写等操作->发生数据给客户端」这个过程是由一个线程(主线程)来完成的</b> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">但是,Redis 程序并不是单线程的,Redis 在启动的时候,是会启动后台线程(BIO)的:
		• <strong>Redis 在 2.6 版本</strong>,会启动 2 个后台线程,分别处理关闭文件、AOF 刷盘这两个任务; 
		• <strong>Redis 在 4.0 版本之后</strong>,新增了一个新的后台线程,用来异步释放 Redis 内存,也就是 lazyfree 线程。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">Redis 为「关闭文件、AOF 刷盘、释放内存」这些任务创建单独的线程来处理,是因为这些任务的操作都是很耗时的,如果把这些任务都放在主线程来处理,那么 Redis 主线程就很容易发生阻塞,这样就无法处理后续的请求了。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">后台线程相当于一个消费者,生产者把耗时任务丢到任务队列中,消费者(BIO)不停轮询这个队列,拿出任务就去执行对应的方法即可。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  #Redis单线程模式是怎样的<span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 单线程模式是怎样的?<br></span><span class = "bq">图中的蓝色部分是一个事件循环,是由主线程负责的,可以看到网络 I/O 和命令处理都是单线程。Redis 初始化的时候,会做下面这几年事情:  
	•首先,调用 epoll_create() 创建一个 epoll 对象和调用 socket() 一个服务端 socket 
	 •然后,调用 bind() 绑定端口和调用 listen() 监听该 socket; 
	  •然后,将调用 epoll_crt() 将 listen socket 加入到 epoll,同时注册「连接事件」处理函数。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">初始化完后,主线程就进入到一个<strong>事件循环函数</strong>,主要会做以下事情: 
		•首先,先调用<strong>处理发送队列函数</strong>,看是发送队列里是否有任务,如果有发送任务,则通过 write 函数将客户端发送缓存区里的数据发送出去,如果这一轮数据没有发生完,就会注册写事件处理函数,等待 epoll_wait 发现可写后再处理 。  
		•接着,调用 epoll_wait 函数等待事件的到来:  
		•如果是 <strong>连接事件</strong>到来,则会调用 <strong>连接事件处理函数</strong>,该函数会做这些事情:调用 accpet 获取已连接的 socket ->调用 epoll_ctr 将已连接的 socket 加入到 epoll -> 注册「读事件」处理函数; 
		•如果是 <strong>读事件</strong>到来,则会调用 <strong>读事件处理函数</strong>,该函数会做这些事情:调用 read 获取客户端发送的数据 -> 解析命令 -> 处理命令 -> 将客户端对象添加到发送队列 -> 将执行结果写到发送缓存区等待发送; 
		•如果是 <strong>写事件</strong>到来,则会调用 <strong>写事件处理函数</strong>,该函数会做这些事情:通过 write 函数将客户端发送缓存区里的数据发送出去,如果这一轮数据没有发生完,就会继续注册写事件处理函数,等待 epoll_wait 发现可写后再处理 。 以上就是 Redis 单线模式的工作方式</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 采用单线程为什么还这么快?<br></span><span class = "bq"><b class="focus1"> 官方使用基准测试的结果是,单线程的 Redis 吞吐量可以达到 10W/每秒; </b>有如下几个原因: 
	•Redis 的大部分操作 都在内存中完成,并且采用了高效的数据结构,因此 Redis 瓶颈可能是机器的内存或者网络带宽,而并非 CPU,既然 CPU 不是瓶颈,那么自然就采用单线程的解决方案了; 
	•Redis 采用单线程模型可以 避免了多线程之间的竞争,省去了多线程切换带来的时间和性能上的开销,而且也不会导致死锁问题。 
	 •Redis 采用了 I/O 多路复用机制处理大量的客户端 Socket 请求, IO 多路复用机制是指一个线程处理多个 IO 流,就是我们经常听到的 select/epoll 机制。简单来说,在 Redis 只运行单线程的情况下,该机制允许内核中,同时存在多个监听 Socket 和已连接 Socket。内核会一直监听这些 Socket 上的连接请求或数据请求。
 一旦有请求到达,就会交给 Redis 线程处理,这就实现了一个 Redis 线程处理多个 IO 流的效果。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 6.0 之前为什么使用单线程?<br></span>
	- <span class = "bq">核心意思是:
	<strong>CPU 并不是制约 Redis 性能表现的瓶颈所在</strong>,更多情况下是受到内存大小和网络 I/O 的限制,所以 Redis 核心网络模型使用单线程并没有什么问题,如果你想要使用服务的多核 CPU,可以在一台服务器上启动多个节点或者采用分片集群的方式。 
 除了上面的官方回答,选择单线程的原因也有下面的考虑。 
 使用了单线程后,可维护性高,多线程模型虽然在某些方面表现优异,但是它却引入了程序执行顺序的不确定性,带来了并发读写的一系列问题, 增加了系统复杂度、同时可能存在线程切换、甚至加锁解锁、死锁造成的性能损耗。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 6.0 之后为什么引入了多线程?<br></span><span class = "bq"><strong>在 Redis 6.0 版本之后,也采用了多个 I/O 线程来处理网络请求,这是因为随着网络硬件的性能提升,Redis 的性能瓶颈有时会出现在网络 I/O 的处理上。 </strong>
 所以为了提高网络请求处理的并行度,Redis 6.0 对于网络请求采用多线程来处理。<strong>但是对于命令执行,Redis 仍然使用单线程来处理</strong>,所以大家<strong>不要误解</strong> Redis 有多线程同时执行命令。 
 Redis 官方表示,<strong>Redis 6.0 版本引入的多线程 I/O 特性对性能提升至少是一倍以上。</strong> 
 Redis 6.0 版本支持的 I/O多线程特性,默认是 I/O 多线程只处理写操作(write client socket),并不会以多线程的方式处理读操作(read client socket)。要想开启多线程处理客户端读请求,就需要把Redis.conf配置文件中的 io-threads-do-reads 配置项设为 yes。 
 <code class="sum" >
 // 读请求也使用 io  多线程
io-threads-do-readsyes
同时,Redis.conf配置文件中提供了IO多线程个数的配置项。
//io-threadsN,表示启用N-1个I/O多线程(主线程也算一个I/O线程)
io-threads4
 </code> 关于线程数的设置,官方的建议是如果为 4 核的 CPU,建议线程数设置为 2 或 3,如果为 8 核 CPU 建议线程数设置为 6,线程数一定要小于机器核数,线程数并不是越大越好。因此,<strong> Redis 6.0 版本之后</strong>,<b class="focus1">Redis 在启动的时候,默认情况下会有 6 个线程: </b>
 • Redis-server :Redis 的主线程,主要负责执行命令; 
 • bio_close_file、bio_aof_fsync、bio_lazy_free : 三个后台线程,分别异步处理关闭文件任务、AOF 刷盘任务、释放内存任务;
 • io_thd_1、io_thd_2、io_thd_3 : 三个 I/O 线程,io-threads 默认是 4 ,所以会启动 3(4-1)个 I/O 多线程,用来分担 Redis 网络 I/O 的压力。</span> <span style="display:block;margin-bottom:2%;"></span>

## Redis-0 新特性：带你 100% 掌握多线程模型 - 掘金
-  #Redis多线程<span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 6.0 新特性：带你 100% 掌握多线程模型 - 掘金<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 6.0</span><span class = "bq"><ol start="1">主要特性如下: <li> 多线程处理网络 IO; </li><li> 客户端缓存; </li><li> 细粒度权限控制(ACL); </li><li>RESP3  协议的使用; 登录首⻚ </li><li> 用于复制的 RDB 文件不在有用,将立刻被删除; </li><li> RDB 文件加载速度更快;</li></ol></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 6.0 之前为什么不使用多线程?<br></span>官方答复:  
	  使用 Redis 时,几乎不存在 CPU 成为瓶颈的情况, Redis 主要受限于内存和网络。 
	  在一个普通的 Linux 系统上,Redis 通过使用 pipelining  每秒可以处理 100 万个请求,所以如果应用程序主要使用 $O(N)$ 或 $O(log^N)$ 的命令,它几乎不会占用太多 CPU。
	   使用了单线程后,可维护性高。多线程模型虽然在某些方面表现优异,但是它却引入了程序执行顺序的不确定性,带来了并发读写的一系列问题,增加了系统复杂度、同时可能存在线程切换、甚至加锁解锁、死锁造成的性能损耗。 
	   Redis 通过 AE 事件模型以及 IO 多路复用等技术,处理性能非常高,因此没有必要使用多线程。
	   **单线程机制让 Redis 内部实现的复杂度大大降低,Hash 的惰性 Rehash、Lpush 等等『线程不安全』的命令都可以无锁进行。** <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 6.0 之前单线程指的是 Redis 只有一个线程干活么?<br></span><span class = "bq">非也,Redis 在处理客户端的请求时,包括获取 (socket 读)、解析、执行、内容返回 (socket 写) 等都由一个顺序串行的主线程处理,这就是所谓的「单线程」。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  #IO模型 <span style="color:#a85d45;border-bottom:2px solid #a85d45;">那 Redis 6.0 为啥要引入多线程呀?<br></span>
	- 随着硬件性能提升,Redis 的性能瓶颈可能出现网络 IO 的读写,也就是:**单个线程处理网络读写的速度跟不上底层网络硬件的速度**。
	- 读写网络的read/write  系统调用占用了Redis 执行期间大部分CPU 时间,瓶颈主要在于网络的 IO 消耗, 优化主要有两个方向: <div class="border2">
	 •提高网络 IO 性能,典型的实现比如使用 DPDK 来替代内核网络栈的方式。
	•使用多线程充分利用多核,提高网络请求读写的并行度,典型的实现比如Memcached 。</div>
	  添加对用户态网络协议栈的支持,需要修改 Redis 源码中和网络相关的部分(例如修改所有的网络收发请求函数),这会带来很多开发工作量。
	   而且新增代码还可能引入新 Bug,导致系统不稳定。 所以,Redis 采用多个 IO 线程来处理网络请求,提高网络请求处理的并行度。
	   **需要注意的是,Redis 多 IO 线程模型只用来处理网络读写请求,对于 Redis 的读写命令,依然是单线程处理。**
	  这是因为,网络处理经常是瓶颈,通过多线程并行处理可提高性能。
	   而继续使用单线程执行读写命令,不需要为了保证 Lua 脚本、事务、等开发多线程安全机制,实现更简单。
	  架构图如下:  <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">主线程与 IO 多线程是如何实现协作呢?<br></span><span class = "bq"><ol start="1">主要流程:  <li> 主线程负责接收建立连接请求,获取 <code  class="cs1">socket</code>  放入全局等待读处理队列;  </li><li>主线程通过轮询将可读 <code  class="cs1">socket</code>  分配给 IO 线程;  </li><li> 主线程阻塞等待 IO 线程读取 <code  class="cs1">socket</code>  完成;  </li><li> 主线程执行 IO 线程读取和解析出来的 Redis 请求命令;  </li><li> 主线程阻塞等待 IO 线程将指令执行结果回写回 <code  class="cs1">socket </code>完毕;  </li><li> 主线程清空全局队列,等待客户端后续的请求。</li> 思路:<strong>将主线程 IO 读写任务拆分出来给一组独立的线程处理,使得多个<code  class="cs1"> socket </code>读写可以并行化,但是 Redis 命令还是主线程串行执行</strong>。</ol></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">如何开启多线程呢?<br></span><span class = "bq">Redis 6.0 的多线程默认是禁用的,只使用主线程。如需开启需要修改 redis.conf 配置文件: <code class="cs1">io-threads-do-reads yes </code>。
      关于线程数的设置,官方有一个建议:
      4 核的机器建议设置为 2 或 3 个线程,
      8核的建议设置为 6 个线程,线程数一定要小于机器核数。
       线程数并不是越大越好,官方认为超过了 8 个基本就没什么意义了。
        另外,开启多线程后,还需要设置线程数,否则是不生效的。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">总结与思考<br></span><span class = "bq">随着互联网的⻜速发展,互联网业务系统所要处理的线上流量越来越大,Redis 的单线程模式会导致系统消耗很多 CPU 时间在网络 I/O 上从而降低吞吐量,要提升 Redis 的性能有两个方向:<ul> <li>优化网络 I/O 模块</li>
	   <li>提高机器内存读写的速度</li></ul>
	   后者依赖于硬件的发展,暂时无解。所以只能从前者下手,网络 I/O 的优化又可以分为两个方向: 
	   <ul><li>零拷⻉技术或者 DPDK 技术</li>
	   <li>利用多核优势</li></ul>
	   <strong>模型缺陷</strong>
	   Redis 的多线程网络模型实际上并不是一个标准的 Multi-Reactors/Master-Workers 模型,Redis 的多线程方案中,I/O 线程任务仅仅是通过 socket 读取客户端请求命令并解析,却没有真正去执行命令。
	 所有客户端命令最后还需要回到主线程去执行,因此对多核的利用率并不算高,而且每次主线程都必须在分配完任务之后忙轮询等待所有 I/O 线程完成任务之后才能继续执行其他逻辑。 
	 在我看来,Redis 目前的多线程方案更像是一个折中的选择:既保持了原系统的兼容性,又能利用多核提升 I/O 性能。</span> <span style="display:block;margin-bottom:2%;"></span>
## redis 多进程_深入理解Redis的持久化机制
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">redis 多进程_深入理解Redis的持久化机制<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  #RDB <span style="color:#a85d45;border-bottom:2px solid #a85d45;">RDB持久化<br></span><span class = "bq">Redis为我们提供了两个命令来实现持久化: 
	 <strong>save</strong> - 由主线程直接进行持久化,会阻塞其它客户端的请求。
	 <strong>bgsave</strong>- 由主线程fork一个子线程来进行持久化,不会阻塞其它客户端的请求。</span> <span style="display:block;margin-bottom:2%;"></span>
	 ![[6af1498b122431ba220ccc39204ab9b4.png]]
		- <span class = "bq"><ol start="(1)、">采用bgsave进行持久化的时候,共有如下几个步骤: <li> <strong>主进程</strong>同步fork子进程(包括子进程从父进程进行⻚表的拷⻉),这一步完成后主进程与子进程⻚表所指向的内存是完全一样的;</li><li> <strong>子进程</strong>执行快照数据的持久化,主线程继续处理读写请求,同时主线程还会监视子进程的持久化过程及状态; </li><li> <strong>主进程</strong>如果发现有新数据写入,就会申请新的内存块,拷⻉被修改的⻚面数据对新的内存块中,从新数据写入的时候开始,主进程与子进程的⻚表指向的内存块就开始逐步分离; </li><li> <strong>子进程</strong>RDB持久化完成,子进程结束。</li></ol></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">整个过程看起来都是异步的,但这个RDB的持久化过程有如下问题<br></span><span class = "bq"><div class="border2"><b class="genre5">问题1</b>、如果Redis实例的内存数据很多,那么会导致主进程的<strong>⻚表很大</strong>,其实⻚表的拷⻉是需要阻塞主进程的,这里会有阻塞主进程的⻛险点,有可能导致Redis实例的性能抖动。
 <b class="genre5">问题2</b>、如果子进程的持久化的过程中,有<strong>大量的⻚面数据被修改</strong>,那么主进程将会重新申请大量的内存块(关于写时复制技术,如果有兴趣的话,可以查阅相关操作系统的书籍),来写入新内容,如果内存不够用,有可能导致Redis实例<strong>内存不足</strong>而出现OOM的情况(如果没有开启swap的话),当然如果开启了的swap,新的数据写入将写入磁盘的swap,这也会引发Redis极大的性能⻛险。
 <b class="genre5">问题3</b>、默认情况下,如果异步的子进程在<strong>持久化</strong>的时候<strong>失败</strong>了,那么<strong>主进程默认情况下将会停止接受所有客户端的写入请求。</strong>
 <b class="genre5">问题4</b>、如果运行Redis实例的<strong>绑定了CPU的逻辑核</strong>,那么<strong>子进程</strong>在<strong>持久化的时候,还会与主进程竞争CPU资源,同样会增加主进程的请求处理延迟。</strong>
 <b class="genre5">问题5</b>、如果Redis实例在某个时刻宕机了,<strong>自最后一次执行RDB持久化的时间点之后的所有写入的新数据都会丢失。</strong>
 <b class="genre5">针对问题1:</b>不建议redis单实例的内存太大,最好不要超过8G内存,如果单机不够存储,可以考虑集群化的解决方案。
 <b class="genre5">针对问题2:</b>通常我们需要运行Redis实例的机器内存留有一定的内存富余,从而避免因为RDB子进程持久化时,子进程的写时复制机制引发的内存资源⻛险。
 <b class="genre5">针对问题3:</b>redis也提供参数来进行配置,如果子进程持久化失败可以不阻止主进程接受写入请求,但这样一来就会导致数据没有办法持久化。
 <b class="genre5">针对问题4:</b>通常来看,有两种解决办法:如果为多超线程CPU,让主进程绑定到一个物理核上,如果是单线程CPU,那么主进程就不要绑核(虽然主进程不绑核,会影响一些性能),还有另外一种办法就是修改Redis的源码,这样的话可以让子进程绑定跟主进程不一样的逻辑核。(关于CPU架构相关的知识,如果你有兴趣,可以参阅CPU架构方面的书籍)。
  <b class="genre5">针对问题5:</b>Redis提供了AOF的持久化机制,可以有效解决大量数据丢失的问题,这个后面的小节中会详细介绍AOF持久化。</div></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">RDB持久化优缺点<br></span><span class = "bq">**优点**:  
		(1)、持久化的过程可以异步化,不阻塞主线程处理客户端读写请求;
		 (2)、RDB持久化可以将rdb数据文件压缩成二进制,这样文件尺寸可以比较小,这样故障恢复的速度会比较快,在做主从复制的时候,还可以节省网络传输开销。
		 **缺点**: 
		  (1)、上面的持久化流程图,我们可以看到,RDB持久化可能会造成部分数据的丢失,任何时刻如果Redis实例crash掉了,那么自最后一次执行RDB持久化开始的时间点之后所有新写入的数据都会丢失。
		  (2)、因为压缩技术跟随redis版本的升级可能存在迭进的现象,并非所有版本的redis都能够兼容旧版本的rdb文件,从而实现正常的数据恢复。
		   (3)、持久化的过程,如果内存数据比较多,仍然会有阻塞⻛险并可能显著增加Redis实例对内存的消耗,极端情况下甚至可能出现Redis直接崩溃掉的情况。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  #AOF <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF持久化<br></span><span class = "bq">从上图分析,对于主进程来说,会有这样一些步骤: (1)、处理客户端的命令; (2)、再将客户端的命令同步写入到aof缓冲中; (3)、然后再进行刷盘操作。 </span> ![[280C978A-1BEE-49E5-B3CF-ACF4936BB7AD.png]]<span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">redis是先执行命令再写aof缓存</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">刷盘策略如何选择<br></span><span class = "bq">选择**no**:也就是redis只管写入aof缓冲,此时调用了内核的write函数,而不会调用fsync函数,因此不主动刷盘,由操作系统负责刷盘,这种方式性能最高,但如果操作系统崩溃或主机掉电,数据丢失不可控。
 选择**always**:每次redis都会主动进行fsync操作,直接数据刷盘,这种方式性能最差,但是不会丢失数据。
 选择**everysec**:redis会创建子线程每秒钟去执行异步的刷盘操作,这种情况下,其实就是在性能与数据安全性所做的trade-off,最多只会丢失1秒的数据。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF重写的子进程<br></span><span class = "bq">因为redis的主进程是以**文本追加**的方式不断的写文件,这种方式的好处就是可以使用磁盘的顺序写降低IO操作导致的性能损耗,但是问题在于这种方式也会产生新的问题,那就是因为很多时候,对某一个key来来回回的写入操作可能达到几百条</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF持久化的问题及应对<br></span><span class = "bq">问题1:fork子进程的时候,这个过程是阻塞的,因为操作系统要拷⻉一些主进程的数据结构给子进程,其中很重要的一项就是⻚表,**对⻚表的拷⻉将完全阻塞主进程**,阻塞的⻓短取决于⻚表的大小,这个问题其实在RDB持久化中也是存在的。
 问题2:大家都知道操作系统分配内存是按⻚为单位进行分配的,以centos 7为例,它默认4k,如果在子进程进行初始重写的过程当中,父进程此时写入了一个bigkey,那么父进程就需要重新申请大块内存(根据写时复制机制,主进程在有⻚写入的情况下需要与子进程进行内存分离),那么这个**大块内存申请过程的耗时将会变**⻓,可能会产阻塞⻛险。如果开启了内存大⻚机制,比如说最小分配单元是2M/⻚的话,那么这个阻塞⻛险会更大一些。
针对问题1:不建议redis单实例的内存太大,最好不要超过8G内存,如果单机不够存储,可以考虑集群化的解决方案。
针对问题2:建议是在Redis实例运行的机器上,我们关闭掉大⻚(Huge Page)分配机制,同时尽可能减少在Redis中存储过多的bigkey,因为bigkey的内存释放也是有⻛险的,由于篇幅的关系,这里先暂不展开,如果你有兴趣,可以自己了解一下。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF优缺点<br></span>
	- **优点**:  
	  (1)、数据安全性更高; 
	  (2)、刷盘策略比较灵活(类似于mysql中redo log的刷盘配置);
	   (3)、持久化的容错性较好,即使失败了,也可以通过redis-check-aof工具进行修复
	   (4)、如果最后一条命令执行的是flushall,只要没有执行AOF重写,拿着原来的aof文件,删除掉最后一条flushall命令就可以恢复数据了。
	    **缺点**: 
	     (1)、相同的内存数据,通常情况下,aof文件要比同等内存数据的的rdb文件要大很多;  
	     (2)、aof重写的过程,跟RDB持久化类似的有阻塞的⻛险; 
	      (3)、主从同步的时候,数据传输慢,数据恢复的速度会比RDB慢很多; (4)、如果配置主进程同步写aof文件(即appendfsync always),性能表现往往会非常糟糕。 <span style="display:block;margin-bottom:2%;"></span>
# 缓存过期删除和缓存淘汰
## Redis 过期删除策略和内存淘汰策略有什么区别？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 过期删除策略和内存淘汰策略有什么区别？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">过期删除策略; Redis 是可以对 key 设置过期时间的,因此需要有相应的机制将已过期的键值对删除,而做这个工作的就是过期键值删除策略。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">过期字典存储在 redisDb 结构<br></span><span class = "bq">每当我们对一个 key 设置了过期时间时,Redis 会把该 key 带上过期时间存储到一个**过期字典**(expires dict)中,也就是说「过期字典」保存了数据库中所有 key 的过期时间。
	<b class="focus1">过期字典数据结构结构如下: </b>  
	•过期字典的 key 是一个指针,指向某个键对象;  
	•过期字典的 value 是一个 long long 类型的整数,这个整数保存了 key 的过期时间;; 字典实际上是哈希表,哈希表的最大好处就是让我们可以用 O(1) 的时间复杂度来快速查找。当我们查询一个 key 时,Redis 首先检查该 key 是否存在于过期字典中:  
	•如果不在,则正常读取键值; 
	•如果存在,则会获取该 key 的过期时间,然后与当前系统时间进行比对,如果比系统时间大,那就没有过期,否则判定该 key 已过期。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">常⻅的三种过期删除策略:  
	 •定时删除; 
	 •惰性删除; 
	 •定期删除;</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">定时删除策略<br></span><span class = "bq">定时删除策略的做法是,**在设置 key 的过期时间时,同时创建一个定时事件,当时间到达时,由事件处理器自动执行 key 的删除操作**。
优点: •可以保证过期 key 会被尽快删除,也就是内存可以被尽快地释放。因此,定时删除对内存是最友好的。
缺点: •在过期 key 比较多的情况下,删除过期 key 可能会占用相当一部分 CPU 时间,在内存不紧张但 CPU 时间紧张的情况下,将 CPU 时间用于删除和当前任务无关的过期键上,无疑会对服务器的响应时间和吞吐量造成影响。所以,定时删除策略对 CPU 不友好。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">惰性删除策略<br></span><span class = "bq">惰性删除策略的做法是,**不主动删除过期键,每次从数据库访问 key 时,都检测 key 是否过期,如果过期则删除该 key**。
			 优点: •因为每次访问时,才会检查 key 是否过期,所以此策略只会使用很少的系统资源,因此,惰性删除策略对 CPU 时间最友好。
			 缺点: •如果一个 key 已经过期,而这个 key 又仍然保留在数据库中, 那么只要这个过期 key 一直没有被访问,它所占用的内存就不会释放,造成了一定的内存空间浪费。所以,惰性删除策略对内存不友好。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">定期删除策略<br></span><span class = "bq">定期删除策略的做法是,**每隔一段时间「随机」从数据库中取出一定数量的 key 进行检查,并删除其中的过期key**。
			优点: •通过限制删除操作执行的时⻓和频率,来减少删除操作对CPU 的影响,同时也能删除一部分过期的数据减少了过期键对空间的无效占用。 
			缺点: 
			•内存清理方面没有定时删除效果好,同时没有惰性删除使用的系统资源少。
			•难以确定删除操作执行的时⻓和频率。如果执行的太频繁,定期删除策略变得和定时删除策略一样,对CPU不友好;如果执行的太少,那又和惰性删除一样了,过期 key 占用的内存不会及时得到释放。</span> <span style="display:block;margin-bottom:2%;"></span>
	- #Redis过期删除策略  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 过期删除策略是什么?<br></span><span class = "bq"><b class="focus1">Redis 选择「惰性删除+定期删除」这两种策略配和使用,</b>不主动删除过期键,每次从数据库访问key时,都检测key是否过期,如果过期则删除该key Redis过期删除策略每隔一段时间「随机」从数据库中取出一定数量的key进行检查,并删除其定期删除中的过期key</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 是怎么实现惰性删除的?<br></span><span class = "bq">Redis 的惰性删除策略由 db.c 文件中的 expireIfNeeded 函数实现; Redis 在访问或者修改 key 之前,都会调用 expireIfNeeded 函数对其进行检查,检查 key 是否过期: 
		 •如果过期,则删除该 key,至于选择异步删除,还是选择同步删除,根据 lazyfree_lazy_expire 参数配置决定(Redis 4.0版本开始提供参数),然后返回 null 客户端; 
		 •如果没有过期,不做任何处理,然后返回正常的键值对给客户端;</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 是怎么实现定期删除的?<br></span><span class = "bq">在 Redis 中,默认每秒进行 10 次过期检查一次数据库,此配置可通过 Redis 的配置文件 redis.conf 进行配置,配置键为 hz 它的默认值是 hz 10。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">特别强调下,每次检查数据库并不是遍历过期字典中的所有 key,而是从数据库中随机抽取一定数量的 key 进行过期检查。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">我查了下源码,定期删除的实现在 expire.c 文件下的 activeExpireCycle 函数中,其中随机抽查的数量由 ``ACTIVE_EXPIRE_CYCLE_LOOKUPS_PER_LOOP`` 定义的,它是写死在代码中的,数值是 20。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">详细说说 Redis 的定期删除的流程:<br></span><span class = "bq"><div class="border0">1. 从过期字典中随机抽取 20 个 key; 
		2. 检查这 20 个 key 是否过期,并删除已过期的 key; 
		3. 如果本轮检查的已过期 key 的数量,超过 5 个(20/4),也就是「已过期 key 的数量」占比「随机抽取 key 的数量」大于 25%,则继续重复步骤 1;如果已过期的 key 比例小于 25%, 则停止继续删除过期 key,然后等待下一轮再检查。
		可以看到,定期删除是一个循环的流程。</div></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">内存淘汰策略<br></span><span class = "bq">当 Redis 的运行内存已经超过 Redis 设置的最大内存之后,则会使用内存淘汰策略删除符合条件的 key,以此来保障 Redis 高效的运行。</span>  ![[Pasted image 20230614160847.png]]	    <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq"> 在配置文件 redis.conf 中,可以通过参数 maxmemory /<bytes/> 来设定最大运行内存,只有在 Redis 的运行内存达到了我们设置的最大运行内存,才会触发内存淘汰策略。 不同位数的操作系统, maxmemory 的默认值是不同的
		•在 64 位操作系统中,maxmemory 的默认值是 0,表示没有内存大小限制,那么不管用户存放多少数据到 Redis 中,Redis 也不会对可用内存进行检查,直到 Redis 实例因内存不足而崩溃也无作为。
		•在 32 位操作系统中,maxmemory 的默认值是 3G,因为 32 位的机器最大只支持 4GB 的内存,而系统本身就需要一定的内存资源来支持运行,所以 32 位操作系统限制最大 3 GB 的可用内存是非常合理的,这样可以避免因为内存不足而导致 Redis 实例崩溃。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">Redis 内存淘汰策略共有八种,这八种策略大体分为「不进行数据淘汰」和「进行数据淘汰」两类策略。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">不进行数据淘汰的策略<br></span><span class = "bq">**noeviction**(Redis3.0之后,默认的内存淘汰策略) :它表示当运行内存超过最大设置内存时,不淘汰任何数据,这时如果有新的数据写入,则会触发 OOM,但是如果没用数据写入的话,只是单纯的查询或者删除操作的话,还是可以正常工作。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">进行数据淘汰的策略<br></span><span class = "bq">针对「进行数据淘汰」这一类策略,又可以细分为「在设置了过期时间的数据中进行淘汰」和「在所有数据范围内进行淘汰」这两类策略。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">在设置了过期时间的数据中进行淘汰<br></span><span class = "bq">•**volatile-random:**随机淘汰设置了过期时间的任意键值; 
			 •**volatile-ttl:**优先淘汰更早过期的键值。
			 •**volatile-lru**(Redis3.0 之前,默认的内存淘汰策略):淘汰所有设置了过期时间的键值中,最久未使用的键值; 
			 •**volatile-lfu**(Redis 4.0 后新增的内存淘汰策略):淘汰所有设置了过期时间的键值中,最少使用的键值;</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">在所有数据范围内进行淘汰<br></span><span class = "bq">
			•**allkeys-random**:随机淘汰任意键值; 
			•**allkeys-lru**:淘汰整个键值中最久未使用的键值; 
			•**allkeys-lfu**(Redis 4.0 后新增的内存淘汰策略):淘汰整个键值中最少使用的键值。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">可以使用 ``config get maxmemory-policy`` 命令,来查看当前 Redis 的内存淘汰策略; noeviction 类型的内存淘汰策略, 它是 Redis 3.0 之后默认使用的内存淘汰策略,表示当运行内存超过最大设置内存时,不淘汰任何数据,但新增操作会报错。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">修改 Redis 内存淘汰策略<br></span><span class = "bq">设置内存淘汰策略有两种方法: 
		•方式一:通过“config set maxmemory-policy /<策略/>”命令设置。它的优点是设置之后立即生效,不需要重启 Redis 服务, **缺点是重启 Redis 之后,设置就会失效**。 
		•方式二:通过修改 Redis 配置文件修改,设置“maxmemorypolicy /<策略/>”,它的优点是重启 Redis 服务后配置不会丢失, **缺点是必须重启 Redis 服务,设置才能生效。**</span> <span style="display:block;margin-bottom:2%;"></span>
	- #LRU算法 <span style="color:#a85d45;border-bottom:2px solid #a85d45;">LRU 算法<br></span><span class = "bq">LRU 全称是 Least Recently Used 翻译为最近最少使用,会选择淘汰最近最少使用的数据。 
	传统 LRU 算法的实现是基于「链表」结构,链表中的元素按照操作顺序从前往后排列,最新操作的键会被移动到表头,当需要内存淘汰时,只需要删除链表尾部的元素即可,因为链表尾部的元素就代表最久未被使用的元素。
	Redis 并没有使用这样的方式实现 LRU 算法,因为传统的 LRU 算法存在两个问题: 
	•需要用链表管理所有的缓存数据,这会带来额外的空间开销; 
	•当有数据被访问时,需要在链表上把该数据移动到头端,如果有大量数据被访问,就会带来很多链表移动操作,会很耗时, 进而会降低 Redis 缓存性能。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 实现的是一种近似 LRU 算法<br></span><span class = "bq"><b class="genre1">实现方式是**在 Redis 的对象结构体中添加一个额外的字段,用于记录此数据的最后一次访问时间**。</b></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">当 Redis 进行内存淘汰时,会使用随机采样的方式来淘汰数据,它是随机取 5 个值(此值可配置),然后淘汰最久没有使用的那个。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">优点<br></span><span class = "bq">•不用为所有的数据维护一个大链表,节省了空间占用; 
		 •不用在每次数据访问时都移动链表项,提升了缓存的性能;</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">问题<br></span><span class = "bq">无法解决缓存污染问题,比如应用一次读取了大量的数据,而这些数据只会被读取这一次,那么这些数据会留存在 Redis 缓存中很⻓一段时间,造成缓存污染。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  #LFU算法<span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 4.0 之后引入了 LFU 算法<br></span><span class = "bq">LFU 全称是 Least Frequently Used 翻译为最近最不常用,LFU 算法是根据数据访问次数来淘汰数据的,它的核心思想是“如果数据过去被访问多次,那么将来被访问的频率也更高”。
 所以, LFU 算法会记录每个数据的访问次数。当一个数据被再次访问时,就会增加该数据的访问次数。这样就解决了偶尔被访问一次之后,数据留存在缓存中很⻓一段时间的问题,相比于 LRU 算法也更合理一些。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 是如何实现 LFU 算法<br></span><span class = "bq">LFU 算法相比于 LRU 算法的实现,多记录了「数据的访问频次」的信息。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 对象头中的 lru 字段,在 LRU 算法下和 LFU 算法下使用方式并不相同。<br></span><span class = "bq">**在 LRU 算法**中,Redis 对象头的 **24** bits 的 lru 字段是用来记录 key 的**访问时间戳**,因此在 LRU 模式下,Redis可以根据对象头中的 lru 字段记录的值,来比较最后一次 key 的访问时间⻓,从而淘汰最久未被使用的 key。
 **在 LFU 算法中**,Redis对象头的 24 bits 的 lru 字段被分成两段来存储,**高 16bit 存储** ldt(Last Decrement Time),**低 8bit 存储** logc(Logistic Counter)。
  •ldt 是用来记录 key 的访问时间戳; 
  h n•logc 是用来记录 key 的访问频次,它的值越小表示使用频率越低,越容易淘汰,每个新加入的 key 的logc 初始值为 5。
   logc 会随时间推移而衰减的。
在每次 key 被访问时,会先对 logc 做一个衰减操作,衰减的值跟前后访问时间的差距有关系,如果上一次访问的时间与这一次访问的时间差距很大,那么衰减的值就越大,<b class="focus1"> 这样实现的 LFU 算法是根据访问频率来淘汰数据的,而不只是访问次数。 </b>访问频率需要考虑 key 的访问是多⻓时间段内发生的。key 的先前访问距离当前时间越⻓,那么这个 key 的访问频率相应地也就会降低,这样被淘汰的概率也会更大。
对 logc 做完衰减操作后,就开始对 logc 进行增加操作,增加操作并不是单纯的 + 1,而是根据概率增加,如果 logc 越大的 key,它的 logc 就越难再增加。
 所以,Redis 在访问 key 时,对于 logc 是这样变化的: 
 1. 先按照上次访问距离当前的时⻓,来对 logc 进行衰减; 
 2. 然后,再按照一定概率增加 logc 的值redis.conf 提供了两个配置项,用于调整 LFU 算法从而控制 logc 的增⻓和衰减: 
        • lfu-decay-time 用于调整 logc 的衰减速度,它是一个以分钟为单位的数值,默认值为1,lfu-decay-time 值越大,衰减越慢; 
	    •lfu-log-factor 用于调整 logc 的增⻓速度,lfu-log-factor 值越大,logc 增⻓越慢</span> <span style="display:block;margin-bottom:2%;"></span>
## Redis数据淘汰算法 
- <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis数据淘汰算法<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">Redis的所有数据都存储在内存中,但是内存是一种有限的资源, 所以为了防止Redis无限制的使用内存,在启动Redis时可以通过配置项 `maxmemory` 来指定其最大能使用的内存容量。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">当Redis使用的内存超过配置的 maxmemory 时,便会触发数据淘汰策略。 Redis提供了多种数据淘汰的策略,如下:  
		• volatile-lru: 最近最少使用算法,从设置了过期时间的键中选择空转时间最⻓的键值对清除掉  
		• volatile-lfu: 最近最不经常使用算法,从设置了过期时间的键中选择某段时间之内使用频次最小的键值对清除掉  
		• volatile-ttl: 从设置了过期时间的键中选择过期时间最早的键值对清除  
		• volatile-random: 从设置了过期时间的键中,随机选择键进行清除  
		• allkeys-lru: 最近最少使用算法,从所有的键中选择空转时间最⻓的键值对清除  
		• allkeys-lfu: 最近最不经常使用算法,从所有的键中选择某段时间之内使用频次最少的键值对清除  
		• allkeys-random: 所有的键中,随机选择键进行删除  
		• noeviction: 不做任何的清理工作,在redis的内存超过限制之后,所有的写入操作都会返回错误;但是读操作都能正常的进行</span> <span style="display:block;margin-bottom:2%;"></span>
			- #LRU算法  <span class = "bq">LRU是 Least Recently Used 的缩写,即最近最少使用,很多缓存系统都使用此算法作为淘汰策略。
最简单的实现方式就是把所有缓存通过一个链表连接起来,新创建的缓存添加到链表的头部,如果有缓存被访问了,就把缓存移动到链表的头部。由于被访问的缓存会移动到链表的头部,所以没有被访问的缓存会随着时间的推移移动的链表的尾部,淘汰数据时只需要从链表的尾部开始即可。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis的LRU算法<br></span><span class = "bq">Redis使用了结构体 robj 来存储缓存对象,而 robj 结构有个名为 lru 的字段,用于记录缓存对象最后被访问的时间,Redis就是以 lru 字段的值作为淘汰依据。</span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span class = "bq">如果配置的淘汰策略是 LRU/LFU/TTL 的话,那么就进入 if 代码块。在 if 代码块里,首先调用 evictionPoolPopulate() 函数选择一些缓存对象样本放置到 EvictionPoolLRU 数组中。
 evictionPoolPopulate() 函数后面会进行分析,现在只需要知道 evictionPoolPopulate() 函数是选取一些缓存对象样本就可以了。 
 获取到缓存对象样本后,还需要从样本中获取最合适的缓存对象进行淘汰,因为在选择样本时会把最合适的缓存对象放置在 EvictionPoolLRU 数组的尾部,所以只需要从 EvictionPoolLRU 数组的尾部开始查找一个不为空的缓存对象即可。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">可以在启动Redis时,通过配置项 maxmemory_policy 来指定要使用的数据淘汰策略。</span> <span style="display:block;margin-bottom:2%;"></span>
# 持久化
## AOF 持久化是怎么实现的？
-  #AOF<span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF 持久化是怎么实现的？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF(Append Only File)<br></span><span class = "bq"><b class="focus1"> 注意只会记录写操作命令, 读操作命令是不会被记录的</b></span> 
	
	```json
		//redis.conf
		appendonly yes //表示是否开启aof持久化 默认no
		appendfilename "appendonly.aof" //aof持久化文件的名称
	```  
	
	 - <span class = "bq">Redis 是先执行写操作命令后,才将该命令记录到 AOF 日志里的,这么做其实有两个好处。</span>
	- <span style="color:#a85d45;border-bottom:2px solid #a85d45;">避免额外的检查开销<br></span><span class = "bq">因为如果先将写操作命令记录到 AOF 日志里,再执行该命令的话, 如果当前的命令语法有问题,那么如果不进行命令语法检查,该错误的命令记录到 AOF 日志里后,Redis 在使用日志恢复数据时,就可能会出错。
 而如果先执行写操作命令再记录日志的话,只有在该命令执行成功后,才将命令记录到 AOF 日志里,这样就不用额外的检查开销,保证记录在 AOF 日志里的命令都是可执行并且正确的。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">不会阻塞当前写操作命令的执行,<br></span><span class = "bq">因为当写操作命令执行成功后,才会将命令记录到 AOF 日志。
	当然,AOF 持久化功能也不是没有潜在⻛险。
 第一个⻛险,执行写操作命令和记录日志是两个过程,那当 Redis 在还没来得及将命令写入到硬盘时,服务器发生宕机了,这个数据就会有丢失的⻛险。
 第二个⻛险,前面说道,由于写操作命令执行成功后才记录到 AOF 日志,所以不会阻塞当前写操作命令的执行,但是可能会给「下一个」命令带来阻塞⻛险。
 因为将命令写入到日志的这个操作也是在主进程完成的(执行命令也是在主进程),也就是说这两个操作是同步的。; </span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">三种写回策略<br></span>
	- ![[redis写入aof日志的过程.png]]
	- <span class = "bq">在 redis.conf 配置文件中的 appendfsync 配置项可以有以下 3 种参数可填:  
	•Always,这个单词的意思是「总是」,所以它的意思是每次写操作命令执行完后,同步将 AOF 日志数据写回硬盘;  
	•Everysec,这个单词的意思是「每秒」,所以它的意思是每次写操作命令执行完后,先将命令写入到 AOF 文件的内核缓冲区,然后每隔一秒将缓冲区里的内容写回到硬盘;  
	•No,意味着不由 Redis 控制写回硬盘的时机,转交给操作系统控制写回的时机,也就是每次写操作命令执行完后,先将命令写入到 AOF 文件的内核缓冲区,再由操作系统决定何时将缓冲区内容写回硬盘。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  这 3 种写回策略都**无法能完美解决「主进程阻塞」和「减少数据丢失」的问题,因为两个问题是对立**的,偏向于一边的话,就会要牺牲另外一边,原因如下: 
	- •**Always** 策略的话,可以最大程度保证**数据不丢失**,但是由于它每执行一条写操作命令就同步将 AOF 内容写回**硬盘**,所以是不可避免会影响主进程的性能; 
	- •**No** 策略的话,是交由操作系统来决定何时将 AOF 日志内容写回硬盘,相比于 Always 策略**性能较好**,但是操作系统写回硬盘的时机是不可预知的,如果 AOF 日志内容没有写回硬盘,一旦服务器宕机,就会**丢失**不定数量的**数据**。 
	- •**Everysec** 策略的话,是**折中**的一种方式,避免了 Always 策略的性能开销,也比 No 策略更能避免数据丢失,当然如果上一秒的写操作命令日志没有写回到硬盘,发生了宕机,这一秒内的数据自然也会丢失。; 
	- <b class="genre1">•如果要**高性能**,就选择 No 策略;  •如果要**高可靠**,就选择 Always 策略;  •如果允许数据丢失一点,但又想性能高,就选择 Everysec 策略。</b>
 <span style="display:block;margin-bottom:2%;"></span>
		 - <span class = "bq">深入到源码后,你就会发现这三种策略只是在控制 fsync() 函数的调用时机。</span> <span style="display:block;margin-bottom:2%;"></span>
		•Always 策略就是每次写入 AOF 文件数据后,就执行 fsync() 函数; 
		•Everysec 策略就会创建一个异步任务来执行 fsync() 函数; 
		•No 策略就是永不执行 fsync() 函数;<span style="display:block;margin-bottom:2%;"></span>
	 <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF 重写机制<br></span><span class = "bq">当 AOF 文件的大小超过所设定的阈值后,Redis 就会启用 AOF 重写机制,来压缩 AOF 文件。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">AOF 重写机制是在重写时,读取当前数据库中的所有键值对,然后将每一个键值对用一条命令记录到「新的 AOF 文件」,等到全部记录完后,就将新的 AOF 文件替换掉现有的 AOF 文件。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">所以,重写机制的妙处在于,尽管某个键值对被多条写命令反复修改,最终也只需要根据这个「键值对」当前的最新状态,然后用一条命令去记录键值对,代替之前记录这个键值对的多条命令,这样就减少了 AOF 文件中的命令数量。最后在重写工作完成后,将新的 AOF 文件覆盖现有的 AOF 文件。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">AOF 重写过程,先重写到新的 AOF 文件,重写失败的话,就直接删除这个文件就好,不会对现有的 AOF 文件造成影响。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF 后台重写<br></span><span class = "bq">但是在触发 AOF 重写时,比如当 AOF 文件大于 64M 时,就会对 AOF 文件进行重写,这时是需要读取所有缓存的键值对数据,并为每个键值对生成一条命令,然后将其写入到新的 AOF 文件,重写完后,就把现在的 AOF 文件替换掉。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  #AOF <span class = "bq">Redis 的重写 AOF 过程是由后台子进程 bgrewriteaof 来完成的,这么做可以达到两个好处: 
		•子进程进行 AOF 重写期间,主进程可以继续处理命令请求,从而避免阻塞主进程; 
		•子进程带有主进程的数据副本(数据副本怎么产生的后面会说),这里使用子进程而不是线程,因为如果是使用线程,多线程之间会共享内存,那么在修改共享内存数据的时候,需要通过加锁来保证数据的安全,而这样就会降低性能。而使用子进程,创建子进程时,父子进程是共享内存数据的,不过这个共享的内存只能以只读的方式,而当父子进程任意一方修改了该共享内存,就会发生「写时复制」,于是父子进程就有了独立的数据副本,就不用加锁来保证数据安全。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">子进程是怎么拥有主进程一样的数据副本的呢?<br></span><span class = "bq">主进程在通过 fork 系统调用生成 bgrewriteaof 子进程时,操作系统会把主进程的「⻚表」复制一份给子进程,这个⻚表记录着虚拟地址和物理地址映射关系,而不会复制物理内存,也就是说,两者的虚拟空间不同,但其对应的物理空间是同一个。
		这样一来,子进程就共享了父进程的物理内存数据了,这样能够节约物理内存资源,⻚表对应的⻚表项的属性会标记该物理内存的权限为只读。
 不过,<b class="genre1">当父进程或者子进程在向这个内存发起写操作时,CPU 就会触发写保护中断,这个写保护中断是由于违反权限导致的,然后操作系统会在「写保护中断处理函数」里进行物理内存的复制,并重新设置其内存映射关系,将父子进程的内存读写权限设置为可读写,最后才会对内存进行写操作,这个过程被称为「写时复制(Copy On Write)」。</b>
 写时复制顾名思义,在发生写操作的时候,操作系统才会去复制物理内存,这样是为了防止 fork 创建子进程时,由于物理内存数据的复制时间过⻓而导致父进程⻓时间阻塞的问题。
 当然,操作系统复制父进程⻚表的时候,父进程也是阻塞中的,不过⻚表的大小相比实际的物理内存小很多,所以通常复制⻚表的过程是比较快的。
 不过,如果父进程的内存数据非常大,那自然⻚表也会很大,这时父进程在通过 fork 创建子进程的时候,阻塞的时间也越久。
 所以,有两个阶段会导致阻塞父进程: 
 <div class="border2">•创建子进程的途中,由于要复制父进程的⻚表等数据结构,阻塞的时间跟⻚表的大小有关,⻚表越大,阻塞的时间也越⻓; 
 •创建完子进程后,如果子进程或者父进程修改了共享数据,就会发生写时复制,这期间会拷⻉物理内存,如果内存越大,自然阻塞的时间也越⻓;</div>
  还有个问题,重写 AOF 日志过程中,如果主进程修改了已经存在 key-value,此时这个 key-value 数据在子进程的内存数据就跟主进程的内存数据不一致了,这时要怎么办呢? 为了解决这种数据不一致问题,Redis 设置了一个 AOF 重写缓冲区,这个缓冲区在创建 bgrewriteaof 子进程之后开始使用。
 在重写 AOF 期间,当 Redis 执行完一个写命令之后,它会同时将这个写命令写入到 「AOF 缓冲区」和 「AOF 重写缓冲区」。
  也就是说,在 bgrewriteaof 子进程执行 AOF 重写期间,主进程需要执行以下三个工作:  
  <div class="border2">•执行客户端发来的命令;  
  •将执行后的写命令追加到 「AOF 缓冲区」;  
  •将执行后的写命令追加到 「AOF 重写缓冲区」</div>
  当子进程完成 AOF 重写工作(扫描数据库中所有数据,逐一把内存数据的键值对转换成一条命令,再将命令记录到重写日志)后,会向主进程发送一条信号,信号是进程间通讯的一种方式,且是异步的。 主进程收到该信号后,会调用一个信号处理函数,该函数主要做以下工作: 
  <div class="border2">•将 AOF 重写缓冲区中的所有内容追加到新的 AOF 的文件中, 使得新旧两个 AOF 文件所保存的数据库状态一致; 
  •新的 AOF 的文件进行改名,覆盖现有的 AOF 文件。 信号函数执行完后,主进程就可以继续像往常一样处理命令了。</div></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">总结<br></span><span class = "bq"><div class="border2">每执行一条写操作命令,就将该命令以追加的方式写入到 AOF 文件,然后在恢复时,以逐一执行命令的方式来进行数据恢复。	<p>Redis 提供了三种将 AOF 日志写回硬盘的策略,分别是 Always、Everysec 和 No,这三种策略在可靠性上是从高到低,而在性能上则是从低到高。</p> <p>随着执行的命令越多,AOF 文件的体积自然也会越来越大,为了避免日志文件过大, Redis 提供了 AOF 重写机制,它会直接扫描数据中所有的键值对数据,然后为每一个键值对生成一条写操作命令,接着将该命令写入到新的 AOF 文件,重写完成后,就替换掉现有的 AOF 日志。重写的过程是由后台子进程完成的,这样可以使得主进程可以继续正常处理命令。</p> <p>用 AOF 日志的方式来恢复数据其实是很慢的,因为 Redis 执行命令由单线程负责的,而 AOF 日志恢复数据的方式是顺序执行日志里的每一条命令,如果 AOF 日志很大,这个「重放」的过程就会很慢了。</p></div></span> <span style="display:block;margin-bottom:2%;"></span>
## RDB 快照是怎么实现的？
-  #RDB<span style="color:#a85d45;border-bottom:2px solid #a85d45;">RDB 快照是怎么实现的？ <br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">•AOF 文件的内容是操作命令; 
	 •RDB 文件的内容是二进制数据。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">RDB 快照就是记录某一个瞬间的内存数据,记录的是实际数据,而 AOF 文件记录的是命令操作的日志,而不是实际的数据。
 因此在 Redis 恢复数据时, RDB 恢复数据的效率会比 AOF 高些, 因为直接将 RDB 文件读入内存就可以,不需要像 AOF 那样还需要额外执行操作命令的步骤才能恢复数据。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">Redis 提供了两个命令来生成 RDB 文件,分别是 save 和 bgsave, 他们的区别就在于是否在「主线程」里执行: 
		•执行了 save 命令,就会在主线程生成 RDB 文件,由于和执行操作命令在同一个线程,所以如果写入 RDB 文件的时间太⻓, 会**阻塞主线程**; 
		•执行了 bgsave 命令,会创建一个子进程来生成 RDB 文件,这样可以**避免主线程**的**阻塞**; RDB 文件的加载工作是在服务器启动时自动执行的,Redis 并没有提供专⻔用于加载 RDB 文件的命令。
		 Redis 还可以通过配置文件的选项来实现每隔一段时间自动执行一次bgsave 命令。
		  Redis 的快照是全量快照,也就是说每次执行快照,都是把内存中的「所有数据」都记录到磁盘中。
		  RDB 快照的缺点,在服务器发生故障时,丢失的数据会比 AOF 持久化的方式更多,因为 RDB 快照是全量快照的方式,因此执行的频率不能太频繁,否则会影响 Redis 性能,而 AOF 日志可以以秒级的方式记录操作命令,所以丢失的数据就相对更少。
		  执行 bgsave 过程中,**Redis 依然可以继续处理操作命令的,也就是数据是能被修改的。**</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">写时复制技术(Copy-OnWrite, COW)<br></span><span class = "bq">执行 bgsave 命令的时候,会通过 fork() 创建子进程,此时子进程和父进程是共享同一片内存数据的,因为创建子进程的时候,会复制父进程的⻚表,但是⻚表指向的物理内存还是一个。
		 只有在发生修改内存数据的情况时,物理内存才会被复制一份。
		这样的目的是为了减少创建子进程时的性能损耗,从而加快创建子进程的速度,毕竟创建子进程的过程中,是会阻塞主线程的。
		 所以,创建 bgsave 子进程后,由于共享父进程的所有内存数据,于是就可以直接读取主线程(父进程)里的内存数据,并将数据写入到 RDB 文件。
		  当主线程(父进程)对这些共享的内存数据也都是只读操作,那么, 主线程(父进程)和 bgsave 子进程相互不影响。
		  但是,<b class="genre1">如果主线程(父进程)要**修改共享数据里的某一块数据**(比如键值对 A)时,就会发生写时复制,于是这块数据的物理内存就会被复制一份(键值对 A'),然后主线程在这个数据副本(键值对 A') 进行修改操作。与此同时,bgsave 子进程可以继续把原来的数据(键值对 A)写入到 RDB 文件。</b> 
		   bgsave 快照过程中,如果主线程修改了共享数据,发生了写时复制后,RDB 快照保存的是原本的内存数据,而主线程刚修改的数据,是没办法在这一时间写入 RDB 文件的,只能交由下一次的 bgsave 快照。
		 在 Redis 执行 RDB 持久化期间,刚 fork 时,主进程和子进程共享同一物理内存,但是途中主进程处理了写操作,修改了共享内存,于是当前被修改的数据的物理内存就会被复制一份。
		 那么极端情况下,如果所有的共享内存都被修改,则此时的内存占用是原先的 2 倍。
		  所以,针对写操作多的场景,我们要留意下快照过程中内存的变化, 防止内存被占满了。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">混合使用 AOF 日志和内存快照<br></span><span class = "bq">当开启了混合持久化时,在 AOF 重写日志时,fork 出来的重写子进程会先将与主线程共享的内存数据<b class="style1">以 RDB 方式写入到 AOF 文件</b>, 然后主线程处理的操作命令会被记录在重写缓冲区里,重写缓冲区里的增量命令会以 AOF 方式写入到 AOF 文件,写入完成后通知主进程将新的含有 RDB 格式和 AOF 格式的 AOF 文件替换旧的的 AOF 文件。
		也就是说,<b class="style1">使用了混合持久化,AOF 文件的前半部分是 RDB 格式的全量数据,后半部分是 AOF 格式的增量数据。</b>
		 重启 Redis 加载数据的时候,由于前半部分是 RDB 内容,这样加载的时候速度会很快。
		  加载完 RDB 的内容后,才会加载后半部分的 AOF 内容,这里的内容是 Redis 后台子进程重写 AOF 期间,主线程处理的操作命令,可以使得数据更少的丢失。</span> <span style="display:block;margin-bottom:2%;"></span>
## Redis持久化RDB和AOF优缺点是什么，怎么实现的？我应该用哪一个？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis持久化RDB和AOF优缺点是什么，怎么实现的？我应该用哪一个？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis的持久化策略<br></span><span class = "bq">RDB:快照形式是直接把内存中的数据保存到一个 dump 文件中,定时保存,保存策略。
 AOF:把所有的对Redis的服务器进行修改的命令都存到一个文件里,命令的集合。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  #RDB <span style="color:#a85d45;border-bottom:2px solid #a85d45;">RDB 持久化<br></span><span class = "bq">默认 Redis 是会以快照 “RDB” 的形式将数据持久化到磁盘的,一个二进 制文件,`dump.rdb`</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">工作原理<br></span><span class = "bq">当 Redis 需要做持久化时,Redis 会 fork 一个子进程,子进程将数据写到磁盘上一个临时 RDB 文件中。当子进程完成写临时文件后,将原来的 RDB 替换掉,这样的好处就是可以 copy-on-write。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">RDB 的优点<br></span><span class = "bq">这种文件非常适合用于进行备份: 比如说,你可以在最近的 24 小时内,每小时备份一次 RDB 文件,并且在每个月的每一天,也备份一个 RDB 文件。 这样的话,即使遇上问题,也可以随时将数据集还原到不同的版本。<b class="style1">RDB 非常适用于灾难恢复(disaster recovery)。</b></span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">RDB 的缺点<br></span><span class = "bq">如果你需要尽量避免在服务器故障时丢失数据,那么 RDB 不适合你。 虽然 Redis 允许你设置不同的保存点(save point) 来控制保存 RDB 文件的频率, 但是, 因为RDB 文件需要保存整个数据集的状态, 所以它并不是一个轻松的操作。 因此你可能会至少 5 分钟才保存一次 RDB 文件。 在这种情况下, 一旦发生故障停机, 你就可能会丢失好几分钟的数据。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  #AOF <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF 持久化<br></span><span class = "bq">使用 AOF 做持久化,每一个写命令都通过write函数追加到 appendonly.aof 中,配置方式:启动 AOF 持久化的方式</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">AOF 就可以做到全程持久化,只需要在配置文件中开启(默认是no),appendonly yes开启 AOF 之后,Redis 每执行一个修改数据的命令,都会把它添加到 AOF 文件中,当 Redis 重启时,将会读取 AOF 文件进行“重放”以恢复到 Redis 关闭前的最后时刻。</span> 
			``` bash
			appendsync yes 
			appendsync always #每次有数据修改发生时都会写入aof文件 
			appendfsync everysec #每秒同步一次，该策略为aof缺省 
			```   
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF 的优点<br></span><span class = "bq">使用 AOF 持久化会让 Redis 变得非常耐久(much more durable):你可以设置不同的 fsync 策略,比如无 fsync ,每秒钟一次 fsync ,或者每次执行写入命令时 fsync 。 **AOF 的默认策略为每秒**钟 **fsync 一**次,在这种配置下,Redis 仍然可以保持良好的性能,并且就算发生故障停机,也最多只会丢失一秒钟的数据( fsync 会在后台线程执行,所以主线程可以继续努力地处理命令请求)。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">AOF 的缺点<br></span><span class = "bq">对于相同的数据集来说,AOF 文件的体积通常要大于 RDB 文件的体积。根据所使用的 fsync 策略,AOF 的速度可能会慢于 RDB。 在一般情况下, 每秒 fsync 的性能依然非常高, 而关闭 fsync 可以让 AOF 的速度和 RDB 一样快, 即使在高负荷之下也是如此。 不过在处理巨大的写入载入时,RDB 可以提供更有保证的最大延迟时间(latency)。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">二者的区别<br></span>

|        二者的区别       |               |
| :-----:  |     : --:    |
| RDB持久化是指在指定的时间间隔内将内存中的数据集快照写入磁盘,实际操作过程是fork一个子进程,先将数据集写入临时文件,写入成功后,再替换之前的文件,用二进制压缩存储。|AOF持久化以日志的形式记录服务器所处理的每一个写、删除操作,查询操作不会记录,以文本的方式记录,可以打开文件看到详细的操作记录。|
|RDB 和 AOF ,我应该用哪一个? |      |
| 如果你非常关心你的数据,但仍然可以承受数分钟以内的数据丢失, 那么你可以只使用 RDB 持久。|AOF 将 Redis 执行的每一条命令追加到磁盘中,处理巨大的写入会降低 Redis 的性能,不知道你是否可以接受。|
|**数据库备份和灾难恢复**:定时生成 RDB 快照(snapshot)非常便于进行数据库备份, 并且 RDB 恢复数据集的速度也要比AOF 恢复的速度要快。| |
|Redis 支持同时开启 RDB 和 AOF,系统重启后,Redis 会优先使用 AOF 来恢复数据,这样丢失的数据会最少。||
<span style="display:block;margin-bottom:2%;"></span>
# Redis高可用
## Redis 你了解 Redis 的三种集群模式吗？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">[Redis] 你了解 Redis 的三种集群模式吗？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 支持三种集群方案<br></span><span class = "bq">主从复制模式 Sentinel(哨兵)模式 Cluster 模式</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">主从复制模式<br></span>
	![[Pasted image 20230615204325.png]]			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">主从复制的作用<br></span><span class = "bq">Redis 提供了复制(replication)功能,可以实现当一台数据库中的数据更新后,自动将更新的数据同步到其他数据库上。
	 在复制的概念中,数据库分为两类,一类是主数据库(master),另一类是从数据库(slave)。主数据库可以进行读写操作,当写操作导致数据变化时会自动将数据同步给从数据库。而从数据库一般是只读的,并接受主数据库同步过来的数据。一个主数据库可以拥有多个从数据库,而一个从数据库只能拥有一个主数据库。
	 <b class="focus1">  总结:引入主从复制机制的目的有两个
	 一个是读写分离,分担 "master" 的读写压力
	 一个是方便做容灾恢复</b> </span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">主从复制原理<br></span><span class = "bq"><ul><li>从数据库启动成功后,连接主数据库,发送 SYNC 命令; <li>主数据库接收到 SYNC 命令后,开始执行 BGSAVE 命令生成 RDB 文件并使用缓冲区记录此后执行的所有写命令;</li> <li>主数据库 BGSAVE 执行完后,向所有从数据库发送快照文件,并在发送期间继续记录被执行的写命令; 从数据库收到快照文件后丢弃所有旧数据,载入收到的快照; </li><li>主数据库快照发送完毕后开始向从数据库发送缓冲区中的写命令; </li><li>从数据库完成对快照的载入,开始接收命令请求,并执行来自主数据库缓冲区的写命令;(从数据库初始化完成) </li><li>主数据库每执行一个写命令就会向从数据库发送相同的写命令,从数据库接收并执行收到的写命令(从数据库初始化完成后的操作)</li> <li>出现断开重连后,2.8之后的版本会将断线期间的命令传给重数据库,增量复制。</li><li> 主从刚刚连接的时候,进行全量同步;全同步结束后,进行增量同步。当然,如果有需要,slave 在任何时候都可以发起全量同步。Redis 的策略是,无论如何,首先会尝试进行增量同步,如不成功,要求从机进行全量同步。</li></ul> </span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">主从复制优缺点<br></span><span class = "bq">主从复制**优点**<ul><li> 支持主从复制,主机会自动将数据同步到从机,可以进行读写分离; </li><li>为了分载 Master 的读操作压力,Slave 服务器可以为客户端提供只读操作的服务,写服务仍然必须由Master来完成; </li><li>Slave 同样可以接受其它 Slaves 的连接和同步请求,这样可以有效的分载 Master 的同步压力; </li><li>Master Server 是以非阻塞的方式为 Slaves 提供服务。所以在 Master-Slave 同步期间,客户端仍然可以提交查询或修改请求;</li><li> Slave Server 同样是以非阻塞的方式完成数据同步。在同步期间,如果有客户端提交查询请求,Redis则返回同步之前的数据; </li></ul> 
			 主从复制**缺点**<ul><li>Redis不具备自动容错和恢复功能,主机从机的宕机都会导致前端部分读写请求失败,需要等待机器重启或者手动切换前端的IP才能恢复(也就是要人工介入); </li><li>主机宕机,宕机前有部分数据未能及时同步到从机,切换IP后还会引入数据不一致的问题,降低了系统的可用性; </li><li>如果多个 Slave 断线了,需要重启的时候,尽量不要在同一时间段进行重启。因为只要 Slave 启动,就会发送sync 请求和主机全量同步,当多个 Slave 重启的时候,可能会导致 Master IO 剧增从而宕机。</li><li> Redis 较难支持在线扩容,在集群容量达到上限时在线扩容会变得很复杂;</li></ul> </span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Sentinel(哨兵)模式<br></span><span class = "bq">第一种主从同步/复制的模式,当主服务器宕机后,需要手动把一台从服务器切换为主服务器,这就需要人工干预,费事费力,还会造成一段时间内服务不可用。这不是一种推荐的方式,更多时候,我们优先考虑哨兵模式。
 哨兵模式是一种特殊的模式,首先 Redis 提供了哨兵的命令,<b class="style1">哨兵是一个独立的进程,作为进程,它会独立运行。其原理是哨兵通过发送命令,等待Redis服务器响应,从而监控运行的多个 Redis 实例。</b></span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">哨兵模式的作用<br></span><span class = "bq">通过发送命令,让 Redis 服务器返回监控其运行状态,包括主服务器和从服务器; 当哨兵监测到 master 宕机,会自动将 slave 切换成 master ,然后通过发布订阅模式通知其他的从服务器,修改配置文件,让它们切换主机; 
			然而一个哨兵进程对Redis服务器进行监控,也可能会出现问题,为此,我们可以使用多个哨兵进行监控。各个哨兵之间还会进行监控,这样就形成了多哨兵模式。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">哨兵模式的工作方式<br></span><span class = "bq"><ul><li>每个Sentinel(哨兵)进程以每秒钟一次的频率向整个集群中的 Master 主服务器,Slave 从服务器以及其他Sentinel (哨兵)进程发送一个 PING 命令。</li><li> 如果一个实例(instance)距离最后一次有效回复 PING 命令的时间超过 down-after-milliseconds 选项所指定的值, 则这个实例会被 Sentinel(哨兵)进程标记为**主观下线**(SDOWN) </li><li>如果一个 Master 主服务器被标记为主观下线(SDOWN),则正在监视这个 Master 主服务器的所有 Sentinel(哨兵)进程要以每秒一次的频率确认 Master 主服务器的确进入了主观下线状态当有足够数量的 Sentinel(哨兵)进程(大于等于配置文件指定的值)在指定的时间范围内确认 Master 主服务器进入了主观下线状态(SDOWN), 则 Master 主服务器会被标记为客观下线(ODOWN) </li><li>在一般情况下, 每个 Sentinel(哨兵)进程会以每 10 秒一次的频率向集群中的所有 Master 主服务器、Slave 从服务器发送 INFO 命令。</li><li> 当 Master 主服务器被 Sentinel(哨兵)进程标记为**客观下线**(ODOWN)时,Sentinel(哨兵)进程向下线的 Master 主服务器的所有 Slave 从服务器发送 INFO 命令的频率会从 10 秒一次改为每秒一次。</li><li> 若没有足够数量的 Sentinel(哨兵)进程同意 Master主服务器下线, Master 主服务器的客观下线状态就会被移除。</li><li> 若 Master 主服务器重新向 Sentinel(哨兵)进程发送 PING 命令返回有效回复,Master主服务器的主观下线状态就会被移除。</li></ul></span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">哨兵模式的优缺点<br></span><span class = "bq">优点: <ul><li>哨兵模式是基于主从模式的,所有主从的优点,哨兵模式都具有。</li><li> 主从可以自动切换,系统更健壮,可用性更高(可以看作自动版的主从复制)。</li></ul>
 缺点: <ul><li>Redis较难支持在线扩容,在集群容量达到上限时在线扩容会变得很复杂。</li></ul></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Cluster 集群模式(Redis官方)<br></span>![[Pasted image 20230616104219.png]] <span class = "bq">在这个图中,每一个蓝色的圈都代表着一个 redis 的服务器节点。它们任何两个节点之间都是相互连通的。客户端可以与任何一个节点相连接,然后就可以访问集群中的任何一个节点。对其进行存取和其他操作。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  #哈希表<span style="color:#a85d45;border-bottom:2px solid #a85d45;">集群的数据分片<br></span><span class = "bq">Redis 集群有16384 个哈希槽,每个 key 通过 CRC16 校验后对 16384 取模来决定放置哪个槽。集群的每个节点负责一部分hash槽,举个例子,比如当前集群有3个节点,那么: <ul><li>  节点 A 包含 0 到 5460 号哈希槽 </li><li>节点 B 包含 5461 到 10922 号哈希槽 </li><li>节点 C 包含 10923 到 16383 号哈希槽;</li></ul> 
			在 Redis 的每一个节点上,都有这么两个东⻄,一个是插槽(slot),它的的取值范围是:0-16383。还有一个就是cluster,可以理解为是一个集群管理的插件。当我们的存取的 Key到达的时候,<b class="genre1">Redis 会根据 CRC16 的算法</b> 得出一个结果,然后把结果对 16384 求余数,这样每个 key 都会对应一个编号在 0-16383 之间的哈希槽,通过这个值,去找到对应的插槽所对应的节点,然后直接自动跳转到这个对应的节点上进行存取操作。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">集群的特点<br></span><span class = "bq">所有的 redis 节点彼此互联(PING-PONG机制),内部使用二进制协议优化传输速度和带宽。
 节点的 fail 是通过集群中超过半数的节点检测失效时才生效。
 客户端与 Redis 节点直连,不需要中间代理层.客户端不需要连接集群所有节点,连接集群中任何一个可用节点即可。</span> <span style="display:block;margin-bottom:2%;"></span>
## 主从复制是怎么实现的？
 -  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">主从复制是怎么实现的？ <br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第一次同步<br></span><span class = "bq">使用 replicaof(Redis 5.0 之前使用 slaveof)命令形成主服务器和从服务器的关系。; </span> ` replicaof <服务器 A 的 IP 地址> <服务器 A 的 Redis 端口号>; # 服务器B执行，b变a到从服务器`
	主从服务器间的第一次同步的过程可分为三个阶段: 
	• 第一阶段是建立链接、协商同步; 
	•第二阶段是主服务器同步数据给从服务器;  
	•第三阶段是主服务器发送新写操作命令给从服务器。<span style="display:block;margin-bottom:2%;"></span>![[Pasted image 20230616173019.png]]
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第一阶段:建立链接、协商同步<br></span><span class = "bq">执行了 replicaof 命令后,从服务器就会给主服务器发送 psync 命令,表示要进行数据同步。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第二阶段:主服务器同步数据给从服务器<br></span><span class = "bq">主服务器会执行 bgsave 命令来生成 RDB 文件,然后把文件发送给从服务器。
		 从服务器收到 RDB 文件后,会先清空当前的数据,然后载入 RDB 文件。
		  <div class="border2">为了保证主从服务器的数据一致性,主服务器在下面这三个时间间隙中将收到的写操作命令,写入到 replication buffer 缓冲区里:  
		  •主服务器生成 RDB 文件期间;  
		  •主服务器发送 RDB 文件给从服务器期间;  
		  •「从服务器」加载 RDB 文件期间;</div></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第三阶段:主服务器发送新写操作命令给从服务器<br></span><span class = "bq">在主服务器生成的 RDB 文件发送完,从服务器收到 RDB 文件后, 丢弃所有旧数据,将 RDB 数据载入到内存。完成 RDB 的载入后, 会回复一个确认消息给主服务器。
 接着,主服务器将 replication buffer 缓冲区里所记录的写操作命令发送给从服务器,从服务器执行来自主服务器 replication buffer 缓冲区里发来的命令,这时主从服务器的数据就一致了。
 至此,主从服务器的第一次同步的工作就完成了。</span> <span style="display:block;margin-bottom:2%;"></span><span style="color:#a85d45;border-bottom:2px solid #a85d45;">分摊主服务器的压力</span>![[Pasted image 20230616180548.png]]
		 Redis 也是一样的，从服务器可以有自己的从服务器，我们可以把拥有从服务器的从服务器当作经理角色，它不仅可以接收主服务器的同步数据，自己也可以同时作为主服务器的形式将数据同步给从服务器，通过这种方式，**主服务器生成 RDB 和传输 RDB 的压力可以分摊到充当经理角色的从服务器**。
		 - <span style="color:#a85d45;border-bottom:2px solid #a85d45;">增量复制</span>
		 在 Redis 2.8 之前,如果主从服务器在命令同步时出现了网络断开又恢复的情况,从服务器就会和主服务器重新进行一次全量复制,很明显这样的开销太大了,必须要改进一波。[](marginnote3app://note/BAD9F4B1-34E9-49A3-BA0B-1AE489F3D876)
		 所以,从 Redis 2.8 开始,网络断开又恢复后,从主从服务器会采用增量复制的方式继续同步,也就是只会把网络断开期间主服务器接收到的写操作命令,同步给从服务器。![[a101e4986c91bedad230db24ca7b5dc8.png]]网络断开后，当从服务器重新连上主服务器时，从服务器会通过 psync 命令将自己的复制偏移量 slave_repl_offset 发送给主服务器，主服务器根据自己的 master_repl_offset 和 slave_repl_offset 之间的差距，然后来决定对从服务器执行哪种同步操作：<div class="border2">- 如果判断出从服务器要读取的数据还在 repl_backlog_buffer 缓冲区里，那么主服务器将采用**增量同步**的方式；<br>- 相反，如果判断出从服务器要读取的数据已经不存在 repl_backlog_buffer 缓冲区里，那么主服务器将采用**全量同步**的方式。</div>
		- **repl_backlog_buffer**，是一个「**环形**」缓冲区，用于主从服务器断连后，从中找到差异的数据；
		- **replication offset**，标记上面那个缓冲区的同步进度，主从服务器都有各自的偏移量，主服务器使用 master_repl_offset 来记录自己「_写_」到的位置，从服务器使用 slave_repl_offset 来记录自己「_读_」到的位置。
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">总结<br></span><span class = "bq">主从复制共有三种模式:全量复制、基于⻓连接的命令传播、增量复制。
 主从服务器第一次同步的时候,就是采用全量复制,此时主服务器会两个耗时的地方,分别是生成 RDB 文件和传输 RDB 文件。为了避免过多的从服务器和主服务器进行全量复制,可以把一部分从服务器升级为「经理⻆色」,让它也有自己的从服务器,通过这样可以分摊主服务器的压力。
 <b class="style1">第一次同步完成后,主从服务器都会维护着一个⻓连接,主服务器在接收到写操作命令后,就会通过这个连接将写命令传播给从服务器, 来保证主从服务器的数据一致性。</b>
 如果遇到网络断开,增量复制就可以上场了,不过这个还跟 repl_backlog_size 这个大小有关系。
 如果它配置的过小,主从服务器网络恢复时,可能发生「从服务器」想读的数据已经被覆盖了,那么这时就会导致主服务器采用全量复制的方式。所以为了避免这种情况的频繁发生,要调大这个参数的值, 以降低主从服务器断开后全量同步的概率。</span> <span style="display:block;margin-bottom:2%;"></span>
## 为什么要有哨兵？
-  #哨兵<span style="color:#a85d45;border-bottom:2px solid #a85d45;">为什么要有哨兵？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	- 哨兵节点主要负责三件事情：**监控、选主、通知**。
	- 所以，我们重点要学习这三件事情：
		- 哨兵节点是如何监控节点的？又是如何判断主节点是否真的故障了？
		- 根据什么规则选择一个从节点切换为主节点？
		- 怎么把新主节点的相关信息通知给从节点和客户端呢？
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">候选者<br></span><span class = "bq">•第一,拿到半数以上的赞成票; 
	 <b class="focus1"> •第二,拿到的票数同时还需要大于等于哨兵配置文件中的 quorum 值。</b></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 1 主 4 从,5 个哨兵 ,quorum 设置为 3,如果 2 个哨兵故障,当主节点宕机时,哨兵能否判断主节点“客观下线”?主从能否自动切换?<br></span><span class = "bq">•**哨兵集群可以判定主节点“客观下线**”。哨兵集群还剩下 3 个哨兵,当一个哨兵判断主节点“主观下线”后,询问另外 2 个哨兵后,有可能能拿到 3 张赞同票,这时就达到了 quorum 的值, 因此,哨兵集群可以判定主节点为“客观下线”。
 •**哨兵集群可以完成主从切换**。当有个哨兵标记主节点为「客观下线」后,就会进行选举 Leader 的过程,因为此时哨兵集群还剩下 3 个哨兵,那么还是可以拿到半数以上(5/2+1=3)的票, 而且也达到了 quorum 值,满足了选举 Leader 的两个条件, 所以就能选举成功,因此哨兵集群可以完成主从切换。; quorum 的值建议设置为哨兵个数的二分之一加1,例如 3 个哨兵就设置 2,5 个哨兵设置为 3,而且哨兵节点的数量应该是奇数。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">主从故障转移操作<br></span><span class = "bq">•第一步:在已下线主节点(旧主节点)属下的所有「从节点」里面,挑选出一个从节点,并将其转换为主节点。
	•第二步:让已下线主节点属下的所有「从节点」修改复制目标,修改为复制「新主节点」;
	•第三步:将新主节点的 IP 地址和信息,通过「发布者/订阅者机制」通知给客户端; 
	•第四步:继续监视旧主节点,当这个旧主节点重新上线时,将它设置为新主节点的从节点;</span> ![[集群主从故障转移过程拓扑图.gif]]<span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">怎么判断从节点之前的网络连接状态不好呢?<br></span><span class = "bq">Redis 有个叫 <code class="cs1">down-after-milliseconds * 10 </code>配置项,其down-aftermilliseconds 是主从节点断连的最大连接超时时间。如果在 downafter-milliseconds 毫秒内,主从节点都没有通过网络联系上,我们就可以认为主从节点断连了。如果发生断连的次数超过了 10 次,就说明这个从节点的网络状况不好,不适合作为新主节点。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">接下来要对所有从节点进行三轮考察:<b class="style1">优先级、复制进度、ID 号</b>。在进行每一轮考察的时候,哪个从节点优先胜出,就选择其作为新主节点。
 •第一轮考察:哨兵首先会根据从节点的优先级来进行排序,优先级越小排名越靠前, 
 •第二轮考察:如果优先级相同,则查看复制的下标,哪个从「主节点」接收的复制数据多,哪个就靠前。
 •第三轮考察:如果优先级和下标都相同,就选择从节点 ID 较小的那个。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">总结<br></span><span class = "bq">Redis 在 2.8 版本以后提供的哨兵(Sentinel)机制,它的作用是实现**主从节点故障转移**。它会监测主节点是否存活,如果发现主节点挂了,它就会选举一个从节点切换为主节点,并且把新主节点的相关信息通知给从节点和客户端。
 哨兵一般是以集群的方式部署,至少需要 3 个哨兵节点,哨兵集群主要负责三件事情:<b class="focus1">**监控**、**选主**、**通知**</b> 。
 哨兵节点通过 Redis 的发布者/订阅者机制,哨兵之间可以相互感知,相互连接,然后组成哨兵集群,同时哨兵又通过 INFO 命令,在主节点里获得了所有从节点连接信息,于是就能和从节点建立连接, 并进行监控了。
 1、<b class="genre1">第一轮投票:判断主节点下线</b> 当哨兵集群中的某个哨兵判定主节点下线(主观下线)后,就会向其他哨兵发起命令,其他哨兵收到这个命令后,就会根据自身和主节点的网络状况,做出赞成投票或者拒绝投票的响应。
 当这个哨兵的赞同票数达到哨兵配置文件中的 quorum 配置项设定的值后,这时主节点就会被该哨兵标记为「客观下线」。
 2、<b class="genre1">第二轮投票:选出哨兵leader </b> 某个哨兵判定主节点客观下线后,该哨兵就会发起投票,告诉其他哨兵,它想成为 leader,想成为 leader 的哨兵节点,要满足两个条件: 
 •第一,拿到半数以上的赞成票; 
 •第二,拿到的票数同时还需要大于等于哨兵配置文件中的 quorum 值。
 3、<b class="genre1">由哨兵 leader 进行主从故障转移</b> 选举出了哨兵 leader 后,就可以进行主从故障转移的过程了。
 该操作包含以下四个步骤: 
 <div class="border2"> •第一步:在已下线主节点(旧主节点)属下的所有「从节点」里面,挑选出一个从节点,并将其转换为主节点,选择的规则: <ul><li>过滤掉已经离线的从节点; </li><li>过滤掉历史网络连接状态不好的从节点; </li><li>将剩下的从节点,进行三轮考察:优先级、复制进度、ID 号。在每一轮考察过程中,如果找到了一个胜出的从节点,就将其作为新主节点。</li></ul> 
 •第二步:让已下线主节点属下的所有「从节点」修改复制目标,修改为复制「新主节点」;
  •第三步:将新主节点的 IP 地址和信息,通过「发布者/订阅者机制」通知给客户端; 
  •第四步:继续监视旧主节点,当这个旧主节点重新上线时,将它设置为新主节点的从节点;</div></span> <span style="display:block;margin-bottom:2%;"></span>
## 集群教程 — Redis 命令参考
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">集群教程 — Redis 命令参考<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">集群简介<br></span><span class = "bq">Redis 集群是一个可以在多个 Redis 节点之间进行数据共享的设施(installation)。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">Redis 集群通过分区(partition)来提供一定程度的可用性(availability): 即使集群中有一部分节点失效或者无法进行通讯, 集群也可以继续处理命令请求。
 Redis 集群提供了以下两个好处: 将数据自动切分(split)到多个节点的能力。
 当集群中的一部分节点失效或者无法进行通讯时, 仍然可以继续处理命令请求的能力。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <code class = "sum">./redis-trib.rb create --replicas 1 127.0.0.1:7000 127.0.0.1:7001 \ 127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005</code> <span style="display:block;margin-bottom:2%;"></span>
## 切片集群：数据增多了，是该加内存还是加实例？.md
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">09 切片集群：数据增多了，是该加内存还是加实例？.md<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">切片集群,也叫分片集群,就是指启动多个 Redis 实例组成一个集群,然后按照一定的规则,把收到的数据划分成多份,每一份用一个实例来保存。回到我们刚刚的场景中,如果把 25GB 的数据平均分成 5 份(当然,也可以不做均分),使用 5 个实例来保存,每个实例只需要保存 5GB 数据。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">如何保存更多数据?<br></span>
		- **纵向扩展**:升级单个 Redis 实例的资源配置,包括增加**内存**容量、增加**磁盘**容量、使用更高配置的 CPU。就像下图中,原来的实例内存是 8GB,硬盘是 50GB,纵向扩展后,内存增加到24GB,磁盘增加到 150GB。 
		- **横向扩展**:横向增加当前 Redis **实例**的**个数**,就像下图中,原来使用 1 个 8GB 内存、50GB 磁盘的实例,现在使用三个相同配置的实例。 ![[Pasted image 20230627153821.png]]span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">纵向扩展的好处是,实施起来简单、直接。</span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span class = "bq">第一个问题是,当使用 RDB 对数据进行持久化时,如果数据量增加,需要的内存也会增加,主线程 fork 子进程时就可能会阻塞(比如刚刚的例子中的情况)。不过,如果你不要求持久化保存Redis 数据,那么,纵向扩展会是一个不错的选择。</span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span class = "bq">第二个问题:纵向扩展会受到硬件和成本的限制。这很容易理解,毕竟, 把内存从 32GB 扩展到 64GB 还算容易,但是,要想扩充到 1TB,就会面临硬件容量和成本上的限制了。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "bq">横向扩展是一个扩展性更好的方案。这是因为,要想保存更多的数据,采用这种方案的话,只用增加 Redis 的实例个数就行了,不用担心单个实例的硬件和成本限制。在面向百万、千万级别的用户规模时,横向扩展的 Redis 切片集群会是一个非常好的选择。</span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span class = "bq">要想把切片集群用起来,我们就需要解决两大问题: <ul><li>数据切片后,在多个实例之间如何分布? </li><li>客户端怎么确定想要访问的数据在哪个实例上?</li></ul> </span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">数据切片和实例的对应分布关系<br></span><span class = "bq"><b class="focus1"> Redis Cluster 方案</b> 采用哈希槽(Hash Slot,接下来我会直接称之为 Slot),来处理数据和实例之间的映射关系。在 Redis Cluster 方案中,一个切片集群共有 16384 个哈希槽,这些哈希槽类似于数据分区,每个键值对都会根据它的 key,被映射到一个哈希槽中。首先根据键值对的 key,按照CRC16 算法计算一个 16 bit 的值;然后,再用这个 16bit 值对 16384 取模,得到 0~16383 范围内的模数,每个模数代表一个相应编号的哈希槽。
				在手动分配哈希槽时,需要把 16384 个槽都分配完,否则 Redis 集群无法正常工作。</span> <span style="display:block;margin-bottom:2%;"></span>
				-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">客户端如何定位数据?<br></span><span class = "bq">一般来说,客户端和集群实例建立连接后,实例就会把哈希槽的分配信息发给客户端。但是,在集群刚刚创建的时候,每个实例只知道自己被分配了哪些哈希槽,是不知道其他实例拥有的哈希槽信息的。
				 客户端收到哈希槽信息后,会把哈希槽信息缓存在本地。当客户端请求键值对时,会先计算键所对应的哈希槽,然后就可以给相应的实例发送请求了。
				在集群中,实例和哈希槽的对应关系并不是一成不变的,最常⻅的变化有两个: <ul><li>在集群中,实例有新增或删除,Redis 需要重新分配哈希槽</li><li> 为了负载均衡,Redis 需要把哈希槽在所有实例上重新分布一遍。</li></ul> Redis Cluster 方案提供了一种重定向机制,所谓的“重定向”,就是指,客户端给一个实例发送数据读写操作时,这个实例上并没有相应的数据,客户端要再给一个新实例发送操作命令。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">小结<br></span><span class = "bq">这节课,我们学习了切片集群在保存大量数据方面的优势,以及基于哈希槽的数据分布机制和客户端定位键值对的方法。
 <b class="genre1">在应对数据量扩容时,虽然增加内存这种纵向扩展的方法简单直接,但是会造成数据库的内存过大,导致性能变慢。Redis 切片集群提供了横向扩展的模式,</b> 也就是使用多个实例,并给每个实例配置一定数量的哈希槽,数据可以通过键的哈希值映射到哈希槽,再通过哈希槽分散保存到不同的实例上。这样做的好处是扩展性好,不管有多少数据,切片集群都能应对。
 另外,集群的实例增减,或者是为了实现负载均衡而进行的数据重新分布,会导致哈希槽和实例的映射关系发生变化,客户端发送请求时,会收到命令执行报错信息。了解了 MOVED 和 ASK 命令,你就不会为这类报错而头疼了。
 我刚刚说过,在 Redis 3.0 之前,Redis 官方并没有提供切片集群方案,但是,其实当时业界已经有了一些切片集群的方案,例如基于客户端分区的 ShardedJedis,基于代理的 Codis、Twemproxy 等。这些方案的应用早于 Redis Cluster 方案,在支撑的集群实例规模、集群稳定性、客户端友好性方面也都有着各自的优势,我会在后面的课程中,专⻔和你聊聊这些方案的实现机制,以及实践经验。这样一来,当你再碰到业务发展带来的数据量巨大的难题时,就可以根据这些方案的特点, 选择合适的方案实现切片集群,以应对业务需求了。</span> <span style="display:block;margin-bottom:2%;"></span>
# 缓存
## 什么是缓存雪崩、击穿、穿透？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">什么是缓存雪崩、击穿、穿透？<br></span> <span style="display:block;margin-bottom:2%;"></span>
- <span class = "bq">Redis 是内存数据库,我们可以将数据库的数据缓存在 Redis 里,相当于数据缓存在内存,内存的读写速度比硬盘快好几个数量级,这样大大提高了系统性能。</span> <span style="display:block;margin-bottom:2%;"></span>
	- #缓存雪崩 <span style="color:#a85d45;border-bottom:2px solid #a85d45;">缓存雪崩<br></span><span class = "bq">通常我们为了保证缓存中的数据与数据库中的数据一致性,会给 Redis 里的数据设置过期时间,当缓存数据过期后,用户访问的数据如果不在缓存里,业务系统需要重新生成缓存,因此就会访问数据库,并将数据更新到 Redis 里,这样后续请求都可以直接命中缓存。
	 当**大量缓存数据**在**同一时间过期(失效)或者 Redis 故障宕机**时,如果此时有大量的用户请求,都无法在 Redis 中处理,于是全部请求都直接访问数据库,从而导致数据库的压力骤增,严重的会造成数据库宕机,从而形成一系列连锁反应,造成整个系统崩溃,这就是**缓存雪崩的**问题; 发生缓存雪崩有两个原因: 
	 •大量数据同时过期;  
	 •Redis 故障宕机;</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">大量数据同时过期<br></span><span class = "bq"><div class="border2">常⻅的应对方法有下面这几种: <ul><li> 均匀设置过期时间; </li><li> 互斥锁; </li><li> 双 key 策略;  </li><li>后台更新缓存;</li></ul> </div></span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">1. 均匀设置过期时间<br></span><span class = "bq">可以在对缓存数据设置过期时间时,给这些数据的过期时间加上一个随机数,这样就保证数据不会在同一时间过期。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">2. 互斥锁<br></span><span class = "bq">**如果发现访问的数据不在 Redis 里,就加个互斥锁,保证同一时间内只有一个请求来构建缓存(**从数据库读取数据**,**再将数据更新到 Redis 里),当缓存构建完成后,再释放锁。未能获取互斥锁的请求,要么等待锁释放后重新读取缓存, 要么就返回空值或者默认值。
 实现互斥锁的时候,最好设置**超时时间**,不然第一个请求拿到了锁, 然后这个请求发生了某种意外而一直阻塞,一直不释放锁,这时其他请求也一直拿不到锁,整个系统就会出现无响应的现象。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">3. 双 key 策略<br></span><span class = "bq">一个是**主 key**,会设置过期时间,一个是**备 key**,不会设置过期,它们只是 key 不一样,但是 value 值是一样的,相当于给缓存数据做了个副本。; 当主 key 过期了,有大量请求获取缓存数据的时候,直接返回备 key 的数据,这样可以快速响应请求。而不用因为 key 失效而导致大量请求被锁阻塞住(采用了互斥锁,仅一个请求来构建缓存),后续再通知后台线程,重新构建主 key 的数据。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">4. 后台更新缓存<br></span><span class = "bq">业务线程不再负责更新缓存,缓存也不设置有效期,而是让缓存“永久有效”,并将更新缓存的工作交由后台线程定时更新。;
			第一种方式,后台线程不仅负责定时更新缓存,而且也负责频繁地检测缓存是否有效,检测到缓存失效了,原因可能是系统紧张而被淘汰的,于是就要⻢上从数据库读取数据,并更新到缓存。; 间隔最好是毫秒级的;
			第二种方式,在业务线程发现缓存数据失效后(缓存数据被淘汰), 通过消息队列发送一条消息通知后台线程更新缓存,后台线程收到消息后,在更新缓存前可以判断缓存是否存在,存在就不执行更新缓存操作;不存在就读取数据库数据,并将数据加载到缓存。这种方式相比第一种方式缓存的更新会更及时,用户体验也比较好。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 故障宕机<br></span><span class = "bq">针对 Redis 故障宕机而引发的缓存雪崩问题,常⻅的应对方法有下面这几种:<ul><li>服务熔断或请求限流机制;  </li><li>构建 Redis 缓存高可靠集群</li></ul> </span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">1. 服务熔断或请求限流机制<br></span><span class = "bq">因为 Redis 故障宕机而导致缓存雪崩问题时,我们可以启动服务熔断机制,**暂停业务应用对缓存服务的访问,直接返回错误**,不用再继续访问数据库,从而降低对数据库的访问压力,保证数据库系统的正常运行,然后等到 Redis 恢复正常后,再允许业务应用访问缓存服务。; 
			为了减少对业务的影响,<b class="genre1">我们可以启用请求限流机制,只将少部分请求发送到数据库进行处理,再多的请求就在入口直接拒绝服务,</b> 等到 Redis 恢复正常并把缓存预热完后,再解除请求限流的机制。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">2. 构建 Redis 缓存高可靠集群<br></span><span class = "bq">服务熔断或请求限流机制是缓存雪崩发生后的应对方案,我们最好通过**主从节点的方式构建 Redis 缓存高可靠集群**。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  #缓存击穿<span style="color:#a85d45;border-bottom:2px solid #a85d45;">缓存击穿<br></span><span class = "bq">缓存中的**某个热点数据过期**了,此时大量的请求访问了该热点数据,就无法从缓存中读取,直接访问数据库,<b class="focus1">数据库很容易就被高并发的请求冲垮</b>,这就是**缓存击穿**的问题。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">两种方案:  
		 •**互斥锁方案**,保证同一时间只有一个业务线程更新缓存,未能获取互斥锁的请求,要么等待锁释放后重新读取缓存,要么就返回空值或者默认值。 
		 •**不给热点数据设置过期时间**,由后台异步更新缓存,或者在热点数据准备要过期前,提前通知后台线程更新缓存以及重新设置过期时间;</span> <span style="display:block;margin-bottom:2%;"></span>
	-  #缓存穿透<span style="color:#a85d45;border-bottom:2px solid #a85d45;">缓存穿透<br></span><span class = "bq">当用户访问的数据,**既不在缓存中,也不在数据库中**,导致请求在访问缓存时,发现缓存缺失,再去访问数据库时,<b class="style1">发现数据库中也没有要访问的数据,没办法构建缓存数据,来服务后续的请求。 </b>那么当有大量这样的请求到来时,数据库的压力骤增,这就是缓存穿透的问题。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span class = "bq">发生一般有这两种情况:  
		•<mark>业务误操作</mark>,缓存中的数据和数据库中的数据都被误删除了, 所以导致缓存和数据库中都没有数据;  
		•<mark>黑客恶意攻击</mark>,故意大量访问某些读取不存在数据的业务;  
		应对缓存穿透的方案,常⻅的方案有三种。 
		•第一种方案,非法请求的限制;  
		•第二种方案,缓存空值或者默认值;  
		•第三种方案,使用布隆过滤器快速判断数据是否存在,避免通过查询数据库来判断数据是否存在;</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第三种方案,使用布隆过滤器快速判断数据是否存在,避免通过查询数据库来判断数据是否存在。<br></span><span class = "bq">我们可以在写入数据库数据时,使用布隆过滤器做个标记,然后在用户请求到来时,业务线程确认缓存失效后,可以通过查询布隆过滤器快速判断数据是否存在,如果不存在,就不用通过查询数据库来判断数据是否存在。
 即使发生了缓存穿透,大量请求只会查询 Redis 和布隆过滤器,而不会查询数据库,保证了数据库能正常运行,Redis 自身也是支持布隆过滤器的。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  #布隆过滤器<span style="color:#a85d45;border-bottom:2px solid #a85d45;">布隆过滤器是如何工作的呢?<br></span><span class = "bq">布隆过滤器由「初始值都为 0 的位图数组」和「 N 个哈希函数」两部分组成。当我们在写入数据库数据时,在布隆过滤器里做个标记, 这样下次查询数据是否在数据库时,只需要查询布隆过滤器,如果查询到数据没有被标记,说明不在数据库中。
 布隆过滤器会通过 3 个操作完成标记: <div class="border2">•第一步,使用 N 个哈希函数分别对数据做哈希计算,得到 N 个哈希值; 
 •第二步,将第一步得到的 N 个哈希值对位图数组的⻓度取模, 得到每个哈希值在位图数组的对应位置。
 •第三步,将每个哈希值在位图数组的对应位置的值设置为 1;
 **查询布隆过滤器说数据存在，并不一定证明数据库中存在这个数据，但是查询到数据不存在，数据库中一定就不存在这个数据**。
 </div></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">总结<br></span><span class = "bq">缓存异常会面临的三个问题:缓存雪崩、击穿和穿透。
 其中,**缓存雪崩**和**缓存击穿**主要原因**是数据不在缓存中**,而导致大量请求访问了数据库,**数据库压力骤增**,容易引发一系列连锁反应,导致系统奔溃。不过,一旦数据被重新加载回缓存,应用又可以从缓存快速读取数据,不再继续访问数据库,数据库的压力也会瞬间降下来。因此,缓存雪崩和缓存击穿应对的方案比较类似。
 而**缓存穿透**主要原因是**数据既不在**缓存**也不在**数据库中。因此,缓存穿透与缓存雪崩、击穿应对的方案不太一样。
 我这里整理了表格,你可以从下面这张表格很好的知道缓存雪崩、击穿和穿透的区别以及应对方案。</span> 

<table>
<thead>
<tr>
<th>缓存异常</th>
<th>产生原因</th>
<th>应对方案</th>
</tr>
</thead>
<tbody><tr>
<td rowspan=2>缓存雪崩</td>
<td>大量数据同时过期</td>
<td>-均匀设置过期时间避免同一时间过期<br>-互斥锁保证同一时间只有一个应用在构建缓存<br>-双key策略主key设置过期时间备key永久主key过期时返回备key的内容<br>-后台更新缓存定时更新、消息队列通知更新</td>
</tr>
<tr>
<td>Redis故障宕机</td>
<td>服务熔断<br>请求限流构建Redis缓存高可靠集群</td>
</tr>
<tr>
<td>缓存击穿</td>
<td>频繁访问的热点数据过期</td>
<td>互斥锁<br>不给热点数据设置过期时间、由后台更新缓存</td>

</tr>
<tr>
<td>缓存穿透</td>
<td>访问的数据既不在缓存也不在数据库</td>
<td>非法请求的限制;<br>缓存空值或者默认值;<br>使用布隆过滤器快速判断数据是否存在;</td>
</tr>
</tbody></table>



<span style="display:block;margin-bottom:2%;"></span>
## 【字节二面】缓存一致性如何保证？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">【字节二面】缓存一致性如何保证？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">简单来说,存储的速度是有区别的,缓存就是把低速存储的结果,临时保存在高速存储,以提高查询效率。<span><div class="border2">方案选型  <br>1. 首先确认**产品上对延迟性的要求**,如果要求极高,且数据有可能变化,别用缓存。  <br>2. 通常来说,方案1就够了。牛牛咨询过4、5个团队,基本都是用方案1,因为使用缓存方案,通常是**读多写少场景**,同时业务上对延迟具有一定的包容性。方案1虽然有一定延时,但比较实用。  <br>3. 如果想增加**更新时的即时性,就选择方案2**,不过一定要注意,针对redis老数据的删除操作不要作为关键路径,影响核心流程。  <br>4. **方案3、方案4均适用于对延时要求比较高的业务**,其区别为前者是推模式,后者是拉模式,而后者具有更强的可靠性, 且无时序性问题。既然都愿意花功夫做处理消息的逻辑,不如一步到位,用方案4</div></span></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">方案一:等待过期,顺其自然<br></span><span class = "bq">使用redis的过期时间,mysql更新时,redis不做处理,等待缓存过期失效,再从mysql拉取缓存。
 这种方式实现简单,但不一致的时间会比较明显,具体由你的业务来配置。如果读请求非常频繁,且过期时间设置较⻓,则会产生很多脏数据。<div class="border3">优点: <br>•redis原生接口,开发成本低,易于实现;  <br>•管理成本低,出问题的概率会比较小。
 不足:<br> •完全依赖过期时间,时间太短容易造成缓存频繁失效,太⻓容易有较⻓时间不一致,对编程者的业务能力,有一定要求。</div></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">方案二:尝试删除,从头再来<br></span><span class = "bq">在方案一的基础上扩展,不光通过key的过期时间兜底,还需要在更新mysql时,同时尝试删除redis,如果删除成功,下次访问该数据,则会直接查询mysql的数据,此时再写入redis,就完成了数据同步。这里为什么说是尝试删除呢?因为有了key本身的过期时间作为保障,最终一致性是一定达成的,主动删除redis数据只是为了减少不一致的时间,但不能让其成为一个关键路径, 影响核心流程。<div class="border3">优点: <br>•相对方案一,达成最终一致性的延迟更小;  <br>•实现成本较低,只是在方案一的基础上,增加了删除逻辑。
 不足: <br>•如果更新mysql成功,删除redis却失败,就退化到了方案一;  <br>•在高并发场景,业务server需要和mysql、redis同时进行连接,这样是损耗双倍的连接资源,容易造成连接数过多的问题。</div></span> <span style="display:block;margin-bottom:2%;"></span>
		-  #消息队列<span style="color:#a85d45;border-bottom:2px solid #a85d45;">方案三:主动更新,信箱投递<br></span><span class = "bq">从被动防守,到主动进攻,在更新mysql之后,redis也要更新, 怎么更新呢?用消息队列!具体来说,是将更新操作交给消息队列,由消息队列保证可靠性,此外再搭建一个消费服务订阅消息队列,来异步更新redis数据。<div class="border3"> 优点: <br>•使用消息队列,就相当于将请求投递至信箱,只要投递成功即完成任务,不用关心结果,实现了进一步解耦;  <br>•消息队列本身具有可靠性,在投递成功的前提下,通过手动提交等手段去消费,可以保证更新操作至少在redis中执行一次。
 不足: <br>•有时序性问题。举个栗子🌰,两台业务服务器在同一时间发出a = 1和a = 5两条请求,若mysql中先执行a=1再执行a=5,则mysql中a的值最终为5;但由于网络传输本身有延迟,所以无法保证两条请求谁先进入消息队列,最终redis的结果可能是1也可能是5,如果是1,mysql和redis中的数据就会产生不一致;  <br>•引入了消息队列,同时要增加消费服务,成本较高;  <br>•依旧有消耗更多客户端连接数的问题。</div></span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">方案四:订阅日志,完全解耦<br></span><span class = "bq">把我们搭建的消费服务作为mysql的一个slave,订阅mysql的binlog日志,解析日志内容,再更新到redis。此方案和业务完全解耦,redis的更新对业务方透明,可以减少心智成本。<div class="border3"> 优点: <br>•在同步服务压力不大情况下,延迟较低;  •和业务完全解耦,在更新mysql时,不需要做额外操作;  <br>•解决了时序性问题,可靠性强。
 缺点:<br> •要单独搭建一个同步服务,并且引入binlog同步机制,成本较大;  <br>•同步服务如果压力比较大,或者崩溃了,那么在较⻓时间内,redis中都是老旧数据。</div></span> <span style="display:block;margin-bottom:2%;"></span>
## 缓存策略：面试中如何回答缓存穿透、雪崩等问题？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">14 缓存策略：面试中如何回答缓存穿透、雪崩等问题？.md<br></span> <span style="display:block;margin-bottom:2%;"></span>
	- #缓存穿透 <span style="color:#a85d45;border-bottom:2px solid #a85d45;">缓存穿透<br></span><span class = "bq">缓存穿透指的是每次查询个别 key 时,key 在缓存系统不命中,此时应用系统就会从数据库中查询,如果数据库中存在这条数据,则获取数据并更新缓存系统。
		 查询缓存中不存在的数据时,每次都要查询数据库。
		**解决缓存穿透的通用方案是**: 给所有指定的 key 预先设定一个默认值,比如空字符串“Null”,当返回这个空字符串“Null”时,我们可以认为这是一个不存在的 key,在业务代码中,就可以判断是否取消查询数据库的操作,或者等待一段时间再请求这个 key。如果此时取到的值不再是“Null”,我们就可以认为缓存中对应的 key 有值了,这就避免出现请求访问到数据库的情况,从而把大量的类似请求挡在了缓存之中。</span>
	-  #缓存击穿  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">缓存并发问题<br></span><span class = "bq">所有**请求**更新的是同一条数据,这不仅会增加数据库的压力,还会因为反复更新缓存而占用缓存资源,这就叫缓存并发。
	**解决缓存并发** 
	1. 首先,客户端发起请求,先从缓存中读取数据,判断是否能从缓存中读取到数据;  
	2. 如果读取到数据,则直接返回给客户端,流程结束;  
	3. 如果没有读取到数据,那么就在 Redis 中使用 setNX 方法设置一个状态位,表示这是一种锁定状态; 
	4. 如果锁定状态设置成功,表示已经锁定成功,这时候请求从数据库中读取数据,然后更新缓存,最后再将数据返回给客户端;  
	5. 如果锁定状态没有设置成功,表示这个状态位已经被其他请求锁定,此时这个请求会等待一段时间再重新发起数据查询;  
	6. 再次查询后发现缓存中已经有数据了,那么直接返回数据给客户端。 这样就能保证在同一时间只能有一个请求来查询数据库并更新缓存系统,其他请求只能等待重新发起查询,从而解决缓存并发的问题。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  #缓存雪崩 <span style="color:#a85d45;border-bottom:2px solid #a85d45;">缓存雪崩<br></span><span class = "bq">对于缓存雪崩问题,我们可以采用两种方案解决。
 **将缓存失效时间随机打散**: 我们可以在原有的失效时间基础上增加一个随机值(比如 1 到 10 分钟)这样每个缓存的过期时间都不重复了,也就降低了缓存集体失效的概率。
 **设置缓存不过期**: 我们可以通过后台服务来更新缓存数据,从而避免因为缓存失效造成的缓存雪崩,也可以在一定程度上避免缓存并发问题。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">缓存热点问题<br></span>
	- <b class="style1"> 怎么设计一个动态缓存热点数据的策略？<br>怎么设计一个缓存操作与业务分离的架构？<br></b> <span class = "bq">那么缓存策略的总体思路:就是通过判断数据最新访问时间来做排名,并过滤掉不常访问的数据, 只留下经常访问的数据,具体细节如下。
 1. 先通过缓存系统做一个**排序队列**(比如存放 1000 个商品),系统会根据商品的访问时间,更新队列信息,越是最近访问的商品排名越靠前。
 2. 同时系统会**定期过滤**掉队列中**排名最后**的 200 个商品,然后再从数据库中**随机读取**出 200 个商品**加入**队列中。
 3. 这样当请求每次到达的时候,会先从队列中获取商品 ID,如果命中,就根据 ID 再从另一个缓存数据结构中读取实际的商品信息,并返回。
 4. 在 Redis 中可以用 **zadd** 方法和 **zrange** 方法来完成排序队列和获取 200 个商品的操作。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">“高内聚,低耦合<br></span><span class = "bq">将缓存操作与业务代码解耦,实现方案上可以通过 MySQL Binlog + Canal + MQ 的方式。</span> <div class="border0">比如用户在应用系统的后台添加一条配置信息，配置信息存储到了 MySQL 数据库中，同时数据库更新了 Binlog 日志数据，接着再通过使用 Canal 组件来获读取最新的 Binlog 日志数据，然后解析日志数据，并通过事先约定好的数据格式，发送到 MQ 消息队列中，最后再由应用系统将 MQ 中的数据更新到 Redis 中，这样就完成了缓存操作和业务代码之间的解耦。</div><span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">总结<br></span> ![[Redis缓存操作.png]]<span style="display:block;margin-bottom:2%;"></span>
## 数据库和缓存如何保证一致性？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">数据库和缓存如何保证一致性？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	- #缓存一致性 <span class = "bq">无论是「先更新数据库,再更新缓存」,还是「先更新缓存, 再更新数据库」,这两个方案都存在并发问题,当两个请求并发更新同一条数据的时候,可能会出现缓存和数据库中的数据不一致的现象。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Cache Aside 策略<br></span><span class = "bq">不更新缓存,而是删除缓存中的数据。然后,到读取数据时,发现缓存中没了数据之后,再从数据库中读取数据,更新到缓存中。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">「读策略」和「写策略」。<br></span><span class = "bq">写策略的步骤: 
	•更新数据库中的数据;  
	•删除缓存中的数据。 
	读策略的步骤:  
	•如果读取的数据命中了缓存,则直接返回数据;  
	•如果读取的数据没有命中缓存,则从数据库中读取数据,然后将数据写入到缓存,并且返回给用户。
	缓存的写入通常要远远快于数据库的写入,
	<b class="focus1"> 「先更新数据库 + 再删除缓存」的方案,是可以保证数据一致性的。</b> 
 而且阿旺为了确保万无一失,还给缓存数据加上了「过期时间」,就算在这期间存在缓存数据不一致,有过期时间来兜底,这样也能达到最终一致。
  问题：在删除缓存(第二个操作)的时候失败了,导致缓存中的数据是旧值。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">小结<br></span><span class = "bq">「先更新数据库,再删除缓存」的方案虽然保证了数据库与缓存的数据一致性,但是每次更新数据的时候,缓存的数据都会被删除,这样会对缓存的命中率带来影响。
 所以,如果我们的业务对缓存命中率有很高的要求,我们可以采用「更新数据库 + 更新缓存」的方案,因为更新缓存并不会出现缓存未命中的情况。
  •在更新缓存前先加个分布式锁,保证同一时间只运行一个请求更新缓存,就会不会产生并发问题了,当然引入了锁后,对于写入的性能就会带来影响。
 •在更新完缓存时,给缓存加上较短的过期时间,这样即时出现缓存不一致的情况,缓存的数据也会很快过期,对业务还是能接受的。;
 针对「先删除缓存,再更新数据库」方案在「读 + 写」并发请求而造成缓存不一致的解决办法是「延迟双删」。
  延迟双删实现的伪代码如下:
  ```shell
#删除缓存 
  redis.delKey(X) 
  #更新数据库 
  db.update(X) 
  #睡眠 
  Thread.sleep(N)  
  #再删除缓存 
  redis.delKey(X) 
```
  请求 A 的睡眠时间就需要大于请求 B 「从数据库读取数据 + 写入缓存」的时间。
   因此,还是比较建议用「先更新数据库,再删除缓存」的方案。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">如何保证「先更新数据库 ,再删除缓存」这两个操作能执行成功?<br></span><span class = "bq">有两种方法:
	•重试机制。
	•订阅 MySQL binlog,再操作缓存。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">重试机制<br></span><span class = "bq">我们可以引入消息队列,将第二个操作(删除缓存)要操作的数据加入到消息队列,由消费者来操作数据。
 •如果应用删除缓存失败,可以从消息队列中重新读取数据,然后再次删除缓存,这个就是重试机制。当然,如果重试超过的一定次数,还是没有成功,我们就需要向业务层发送报错信息了。
 •如果删除缓存成功,就要把数据从消息队列中移除,避免重复操作,否则就继续重试。</span> ![[Pasted image 20230706111437.png]]<span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">订阅 MySQL binlog,再操作缓存<br></span><span class = "bq">「先更新数据库,再删缓存」的策略的第一步是更新数据库,那么更新数据库成功,就会产生一条变更日志,记录在 binlog 里。
 于是我们就可以通过订阅 binlog 日志,拿到具体要操作的数据,然后再执行缓存删除,阿里巴巴开源的 Canal 中间件就是基于这个实现的。
 Canal 模拟 MySQL 主从复制的交互协议,把自己伪装成一个 MySQL 的从节点,向 MySQL 主节点发送 dump 请求,MySQL 收到请求后,就会开始推送 Binlog 给 Canal,Canal 解析 Binlog 字节流之后,转换为便于读取的结构化数据,供下游程序订阅使用。
  如果要想保证「先更新数据库,再删缓存」策略第二个操作能执行成功,我们可以使用「消息队列来重试缓存的删除」,或者「订阅 MySQL binlog 再操作缓存」,这两种方法有一个共同的特点, 都是采用异步操作缓存。</span> ![[Pasted image 20230706113645.png]]<span style="display:block;margin-bottom:2%;"></span>
# 分布式锁
## Redis分布式锁，你用对了吗？
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis分布式锁，你用对了吗？<br></span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">分布式锁是什么?<br></span><span class = "bq">分布式锁就是分布式场景下的锁,比如多台不同机器上的进程,去**竞争同一项资源**,就是分布式锁。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">分布式锁有哪些特性?<br></span><span class = "bq">
	•<font color=#e68a00>互斥性</font>:锁的目的是获取资源的使用权,所以只让一个竞争者持有锁,这一点要尽可能保证; 
	•<font color=#e68a00>安全性</font>:避免死锁情况发生。当一个竞争者在持有锁期间内,由于意外崩溃而导致未能主动解锁,其持有的锁也能够被正常释放,并保证后续其它竞争者也能加锁; 
	•<font color=#e68a00>对称性</font>:同一个锁,加锁和解锁必须是同一个竞争者。不能把其他竞争者持有的锁给释放了,这又称为锁的可重入性。 
	•<font color=#e68a00>可靠性</font>:需要有一定程度的异常处理能力、容灾能力。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">最简化版本<br></span><span class = "bq">setnx key value; 如果key不存在,则会将key设置为value,并返回1;如果key存在,不会有任务影响,返回0。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">最简化版本有一个问题<br></span><span class = "bq">如果获取锁的服务挂掉了,那么锁就一直得不到释放,就像石沉大海,杳无音信。所以,我们需要一个超时来兜底。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  #原子操作<span style="color:#a85d45;border-bottom:2px solid #a85d45;">原子操作<br></span><span class = "bq">`set key value nx ex seconds` nx表示具备setnx特定,ex表示增加了过期时间,最后一个参数就是过期时间的值。</span>![[Pasted image 20230706154204.png]]<span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "focus1">Redis➕ Lua,可以说是专⻔为解决原子问题而生。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "focus1">分布式锁需要满足谁申请谁释放原则,不能释放别人的锁,也就是说,分布式锁,是要有归属的。</span> <span style="display:block;margin-bottom:2%;"></span>
			-  <span class = "focus1">分布式锁的前三个特性:对称性、安全性、可靠性,就满足了。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">主从容灾<br></span><span class = "bq">为Redis配置从节点,当主节点挂了, 用从节点顶包。; 通过增加从节点的方式,虽然一定程度解决了单点的容灾问题, 但并不是尽善尽美的,由于同步有时延,Slave可能会损失掉部分数据,分布式锁可能失效,这就会发生短暂的多机获取到执行权限。</span> 
	- #RedLock<span style="color:#a85d45;border-bottom:2px solid #a85d45;">多机部署<br></span>如果对一致性的要求高一些，可以尝试多机部署，比如Redis的RedLock，大概的思路就是多个机器，通常是奇数个，达到一半以上同意加锁才算加锁成功，这样，可靠性会向ETCD靠近。
	- <span class = "bq">客户端会执行以下操作：
	 1.向5个Redis申请加锁; 
	 2.只要超过一半,也就是3个Redis返回成功,那么就是获取到了锁。如果超过一半失败,需要向每个Redis发送解锁命令; 
	 3.由于向5个Redis发送请求,会有一定时耗,所以锁剩余持有时间,需要减去请求时间。这个可以作为判断依据,如果剩余时间已经为0,那么也是获取锁失败; 
	 4.使用完成之后,向5个Redis发送解锁请求。
	 单点Redis的所有手段,这种多机模式都可以使用,比如为每个节点配置哨兵模式,由于加锁是一半以上同意就成功,那么如果单个节点进行了主从切换,单个节点数据的丢失,就不会让锁失效了。这样增强了可靠性。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span class = "bq">分布式系统中的三大困境(简称NPC)</span> ,没有完全可靠的分布式锁<span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">N:Network Delay(网络延迟)<br></span><span class = "bq">当分布式锁获得返回包的时间过⻓,此时可能虽然加锁成功,但是已经时过境迁,锁可能很快过期。RedLock算了做了些考量, 也就是前面所说的锁剩余持有时间,需要减去请求时间,如此一来,就可以一定程度解决网络延迟的问题。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">P:Process Pause(进程暂停)<br></span><span class = "bq">比如发生GC,获取锁之后GC了,处于GC执行中,然后锁超时。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">C:Clock Drift(时钟漂移)<br></span><span class = "bq">我们直接假设5台机器都发生了时钟漂移,锁瞬间过期了。这时候竞争者B拿到了锁,此时A和B拿到了相同的执行权限。</span> <span style="display:block;margin-bottom:2%;"></span>
## 分布式系统中，如何回答锁的实现原理？.md
-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">06 分布式系统中，如何回答锁的实现原理？.md<br></span> <span style="display:block;margin-bottom:2%;"></span> #面试模拟 <div class="border3">面试官模拟了系统秒杀的场景：为了防止商品库存超售，在并发场景下用到了分布式锁的机制，做商品扣减库存的串行化操作。然后问你：“你如何实现分布式锁？”你该怎么回答呢？<br>
 可选方案有很多，比如：<br>基于关系型数据库 MySQL 实现分布式锁；<br>基于分布式缓存 Redis 实现分布式锁；<br>你：“可以基于 Redis 的 setnx 命令来实现分布式锁。” 面试官：“当拿到锁的服务挂掉，如何防止死锁？” 你：“可以为锁设置一个过期时间。” 面试官：“那如何保证加锁和设置过期时间是原子操作？”....<br><font color=#bbc1bf>如果面试官觉得你回答问题的思路清晰有条理，给出的实现方案也可以落地，并且满足你的业务场景，那么他会认可你具备初中级研发工程师该具备的设计能力，但不要高兴得太早。</font></div>
	 - “如果让借助第三方组件，你怎么设计分布式锁？” 这背后涉及了分布式锁的底层设计逻辑，是你需要掌握的。
	 - 02 讲我提到，在给出方案之前，你要明确待解决的问题点是什么。虽然你可以借助数据库 DB、Redis 和 ZooKeeper 等方式实现分布式锁，但要设计一个分布式锁，就需要明确分布式锁经常出现哪些问题,以及如何解决。
	 - <span class = "bq">**可用问题**:无论何时都要保证锁服务的可用性(这是系统正常执行锁操作的基础)。
 **死锁问题**:客户端一定可以获得锁,即使锁住某个资源的客户端在释放锁之前崩溃或者网络不可达(这是避免死锁的设计原则)。
 **脑裂问题**:集群同步时产生的数据不一致,导致新的进程有可能拿到锁,但之前的进程以为自己还有锁,那么就出现两个进程拿到了同一个锁的问题。</span> <span style="display:block;margin-bottom:2%;"></span>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">基于关系型数据库实现分布式锁<br></span><span class = "bq">基于关系型数据库（如 MySQL） 来实现分布式锁是任何阶段的研发同学都需要掌握的，做法如下：先查询数据库是否存在记录,为了防止<b class="focus1">幻读取(幻读取:事务 A 按照一定条件进行数据读取, 这期间事务 B 插入了相同搜索条件的新数据,事务 A 再次按照原先条件进行读取时,发现了事务B 新插入的数据 )</b> 通过数据库行锁 select for update 锁住这行数据,然后将查询和插入的 SQL 在同一个事务中提交。
	不过你要注意，基于 MySQL 行锁的方式会出现交叉死锁，比如事务 1 和事务 2 分别取得了记录 1 和记录 2 的排它锁，然后事务 1 又要取得记录 2 的排它锁，事务 2 也要获取记录 1 的排它锁，那这两个事务就会因为相互锁等待，产生死锁。
	你可以通过“超时控制”解决交叉死锁的问题，但在高并发情况下，出现的大部分请求都会排队等待，所以性能上存在缺陷，通常会延伸出下面两个问题。</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">数据库的事务隔离级别<br></span><span class = "bq">数据库的四种隔离级别从低到高分别是：<ul><li>读未提交(READ UNCOMMITTED); </li><li>读已提交(READ COMMITTED); </li><li>可重复读(REPEATABLE READ); </li><li>可串行化(SERIALIZABLE)</li></ul> **所以数据库隔离级别越高，系统的并发性能就越差。**</span> <span style="display:block;margin-bottom:2%;"></span>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">基于乐观锁的方式实现分布式锁<br></span><span class = "bq">在数据库层面,select for update 是悲观锁,会一直阻塞直到事务提交,所以为了不产生锁等待而消耗资源,你可以基于乐观锁的方式来实现分布式锁,比如基于版本号的方式,首先在数据库增加一个 int 型字段 ver,然后在 SELECT 同时获取 ver 值,最后在 UPDATE 的时候检查 ver 值是否为与第 2 步或得到的版本值相同。
```sql
## SELECT 同时获取 ver 值

select amount, old_ver from order where order_id = xxx

## UPDATE 的时候检查 ver 值是否与第 2 步获取到的值相同

update order set ver = old_ver + 1, amount = yyy where order_id = xxx and ver = old_ver
```
此时，如果更新结果的记录数为1，就表示成功，如果更新结果的记录数为 0，就表示已经被其他应用更新过了，需要做异常处理。</span> <span style="display:block;margin-bottom:2%;"></span>
	- <span style="color:#a85d45;border-bottom:2px solid #a85d45;"> 基于分布式缓存实现分布式锁</span><br> #setnx <span class = "bq">基于缓存实现的分布式锁，就是将数据仅存放在系统的内存中，不写入磁盘，从而减少 I/O 读写。**接下来，我以 Redis 为例讲解如何实现分布式锁。**在加锁的过程中，实际上就是在给 Key 键设置一个值，为避免死锁，还要给 Key 键设置一个过期时间。

```sql
SET lock_key unique_value NX PX 10000
```

- lock_key 就是 key 键；
- unique_value 是客户端生成的唯一的标识；
- NX 代表只在 lock_key 不存在时，才对 lock_key 进行设置操作；
- PX 10000 表示设置 lock_key 的过期时间为 10s，这是为了避免客户端发生异常而无法释放锁。
而解锁的过程就是将 lock_key 键删除,实现方式可以通过 lua 脚本判断 unique_value 是否为加锁客户端。
选用 Lua 脚本是为了保证解锁操作的原子性。因为 Redis 在执行 Lua 脚本时，可以以原子性的方式执行，从而保证了锁释放操作的原子性。

```kotlin
// 释放锁时，先比较 unique_value 是否相等，避免锁的误释放

if redis.call("get",KEYS[1]) == ARGV[1] then

    return redis.call("del",KEYS[1])

else

    return 0

end
```

以上，就是基于 Redis 的 SET 命令和 Lua 脚本在 Redis 单节点上完成了分布式锁的加锁、解锁</span><br>
	-  <span class="bq">**你不能仅停留在操作上，因为这并不能满足应对面试需要掌握的知识深度，** 所以你还要清楚以下3个问题。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">基于 Redis 实现分布式锁的优缺点<br></span><span class="bq">基于数据库实现分布式锁的方案来说,基于缓存实现的分布式锁主要的优点主要有三点。
 1. 性能高效(这是选择缓存实现分布式锁最核心的出发点)。
 2. 实现方便。很多研发工程师选择使用 Redis 来实现分布式锁,很大成分上是因为 Redis 提供了setnx 方法,实现分布式锁很方便。但是需要注意的是,在 Redis2.6.12 的之前的版本中,由于加锁命令和设置锁过期时间命令是两个操作(不是原子性的),当出现某个线程操作完成setnx 之后,还没有来得及设置过期时间,线程就挂掉了,就会导致当前线程设置 key 一直存在,后续的线程无法获取锁,最终造成死锁的问题,<b class="genre1">所以要选型 Redis 2.6.12 后的版本或通过Lua 脚本执行加锁和设置超时时间</b>(Redis 允许将 Lua 脚本传到 Redis 服务器中执行, 脚本中可以调用多条 Redis 命令,并且 Redis 保证脚本的原子性)。
 3. 避免单点故障(因为 Redis 是跨集群部署的,自然就避免了单点故障)。
当然,基于 Redis 实现分布式锁也存在缺点,主要是不合理设置超时时间,以及 Redis 集群的数据同步机制,都会导致分布式锁的不可靠性。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">如何合理设置超时时间<br></span><span class="bq">如果锁的超时时间设置过⻓,会影响性能,如果设置的超时时间过短,有可能业务阻塞没有处理完成,<b class="genre1">能否合理设置超时时间,是基于缓存实现分布式锁很难解决的一个问题。</b>
		- 你可以基于续约的方式设置超时时间:先给锁设置一个超时时间,然后启动一个守护线程,让守护线程在一段时间后,重新设置这个 #锁的超时时间 。实现方式就是:写一个守护线程,然后去判断锁的情况,当锁快失效的时候,再次进行续约加锁,当主线程执行完成后,销毁续约锁即可。
		-  不过这种方式实现起来相对复杂,我建议你结合业务场景进行回答,所以针对超时时间的设置,要站在实际的业务场景中进行衡量。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 如何解决集群情况下分布式锁的可靠性?<br></span><span class="bq">由于 Redis 集群数据同步到各个节点时是异步的,如果在 Redis 主节点获取到锁后,在没有同步到其他节点时,Redis 主节点宕机了,此时新的 Redis 主节点依然可以获取锁,所以多个应用服务就可以同时获取到锁。</span> <br>
			- #RedLock  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redlock 原理<br></span><span class="bq">为了避免 Redis 实例故障导致锁无法工作的问题,Redis 的开发者 Antirez 设计了分布式锁算法Redlock。Redlock 算法的基本思路,<b class="genre1">是让客户端和多个独立的 Redis 实例依次请求申请加锁,如果客户端能够和半数以上的实例成功地完成加锁操作,那么我们就认为,客户端成功地获得分布式锁,否则加锁失败。</b>
			- 我们假设目前有 N 个独立的 Redis 实例, 客户端先按顺序依次向 N 个 Redis 实例执行加锁操作。
 这里的加锁操作和在单实例上执行的加锁操作一样,但是需要注意的是,Redlock 算法设置了加锁的超时时间,为了避免因为某个 Redis 实例发生故障而一直等待的情况。
 当客户端完成了和所有 Redis 实例的加锁操作之后,如果有超过半数的 Redis 实例成功的获取到了锁,并且总耗时没有超过锁的有效时间,那么就是加锁成功。</span> <br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">总结<br></span><span class="bq">分布式锁是解决多个进程同时访问临界资源的常用方法,在分布式系统中非常普遍,常⻅的实现方式是基于数据库,基于 Redis。在同等服务器配置下,Redis 的性能是最好的,数据库最差。
对于分布式锁,你要从“解决可用性、死锁、脑裂”等问题为出发点来展开回答各分布式锁的实现方案的优缺点和适用场景。 另外,在设计分布式锁的时候,为了解决可用性、死锁、脑裂等问题, 一般还会再考虑一下锁的四种设计原则。</span> 
 <ul><li><strong>互斥性</strong>:即在分布式系统环境下,对于某一共享资源,需要保证在同一时间只能一个线程或进程对该资源进行操作。</li><li><strong>高可用</strong>:也就是可靠性,锁服务不能有单点⻛险,要保证分布式锁系统是集群的,并且某一台机器锁不能提供服务了,其他机器仍然可以提供锁服务。</li><li><strong>锁释放</strong>:具备锁失效机制,防止死锁。即使出现进程在持有锁的期间崩溃或者解锁失败的情况,也能被动解锁,保证后续其他进程可以获得锁。</li><li><strong>可重入</strong>:一个节点获取了锁之后,还可以再次获取整个锁资源。</li></ul> <br>
## Redis 分布式锁的正确实现原理演化历程与 Redisson 实战总结 - 文章详情

 -  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 分布式锁的正确实现原理演化历程与 Redisson 实战总结 - 文章详情<br></span><br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">进入正文之前,我们先带着问题去思考<br></span><span class="bq">什么时候需要分布式锁? <br> 加、解锁的代码位置有讲究么? <br>  如何避免出现锁再也无法删除?  <br> 超时时间设置多少合适呢?  <br> 如何避免锁被其他线程释放如何实现重入锁? <br>  主从架构会带来什么安全问题?(略，不打算读)  <br> 什么是 Redlock Redisson 分布式锁佳实战看⻔狗实现原理</span> <br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">什么时候用分布式锁?<br></span><span class="bq">当并发去读写一个【共享资源】的时候,我们为了保证数据的正确,需要控制同一时刻只有一个线程访问。</span> <br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">分布式锁入⻔<br></span><span class="bq">分布式锁应该满足哪些特性?
<b class="focus1">互斥; 在任何给定时刻，只有一个客户端可以持有锁；
无死锁; 任何时刻都有可能获得锁,即使获取锁的客户端崩溃;
容错; 只要大多数 Redis 的节点都已经启动,客户端就可以获取和释放锁。</b>
使用 SETNX key value  命令是实现「互斥」特性; SET if Not eXists; 释放锁 DEL lockKey; 这个方案存在一个存在造成锁无法释放的问题,造成该问题的场景如下: 
1. 客户端所在节点崩溃,无法正确释放锁;  
2. 业务逻辑异常,无法执行 DEL 指令。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">超时设置<br></span> #原子操作 <span class="bq">「加锁」、「设置超时」是两个命令,他们不是原子操作。
 如果出现只执行了一条,第二条没机会执行就会出现「超时时间」设置失败,依然出现锁无法释放。
`SET resource_name random_value NX PX 30000;` 
Redis 2.6.X 之后,官方拓展了 SET  命令的参数,满足了当 key 不存在则设置 value,同时设置超时时间的语义,并且满足原子性。
NX:表示只有 resource_name  不存在的时候才能 SET  成功,从而保证只有一个客户端可以获得锁; 
PX 30000:表示这个锁有一个 30 秒自动过期时间。;
释放了不是自己加的锁;  GET + DEL  指令组合而成的,这里又会涉及到原子性问题。;
可以通过 Lua  脚本来实现,这样判断和删除的过程就是原子操作了。
``` lua 
// 获取锁的 value 与 ARGV[1] 是否匹配,匹配则执行 del 
if redis.call("get",KEYS[1]) == ARGV[1] then     
  return redis.call("del",KEYS[1]) 
else     
  return end
```
</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">正确设置锁超时<br></span><span class="bq">一般要根据在测试环境多次测试,然后压测多轮之后,比如计算出平均执行时间 200 ms。; 锁的超时时间就放大为平均执行时间的 3~5 倍。; 如果锁的操作逻辑中有网络 IO 操作、JVM FullGC 等,线上的网络不会总一帆⻛顺,我们要给网络抖动留有缓冲时间。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">完美的方案<br></span><span class="bq">让获得锁的线程开启一个守护线程,用来给快要过期的锁「续航」。
		加锁的时候设置一个过期时间 ,同时客户端开启一个「 守 护 线 程 」 ,定时去检测这个锁的失效时间。
 如果快要过期,但是业务逻辑还没执行完成,自动对这个锁进行续期,重新设置过期时间。</span> <br>
		- #Redisson实现可重锁 <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redisson<br></span><span class="bq">在使用分布式锁时,它就采用了「自动续期」的方案来避免锁过期,这个守护线程我们一般也把它叫做「看⻔狗」线程。
<div class="border2">1. 通过 SET lock_resource_name random_value NX PX expire_time ,同时启动守护线程为快要过期但还没执行完的客户端的锁续命; <br>2. 客户端执行业务逻辑操作共享资源; <br>3. 通过 Lua  脚本释放锁,先 get 判断锁是否是自己加的,再执行 DEL 。</div> </span>
	**加解锁代码位置有讲究**

```java
try {
	   // 上锁        
	   redisLock.lock();        
	   // 处理业务        
	  .. .
      } catch (Exception e) { 
        e.printStackTrace();
      } finally { 
         // 释放锁
        redisLock.unlock(); 
      }
``` 
-  <span class="bq"><b class="focus1"> 实现可重入锁</b>
			当一个线程执行一段代码成功获取锁之后,继续执行时,又遇到加锁的代码,可重入性就就保证线程能继续执行,而不可重入就是需要等待锁释放之后,再次获取锁成功,才能继续往下执行。
Redis Hash 可重入锁; Redisson 类库就是通过 Redis Hash 来实现可重入锁; 我们可以使用 Redis hash 结构实现,key 表示被锁的共享资源, hash 结构的 fieldKey 的value 则保存加锁的次数。
通过 Lua 脚本实现原子性,假设 KEYS1 = 「lock」, ARGV「1000,uuid」:</span> 


```lua
---- 1 代表 true  
----  代表 false  
if (redis.call('exists', KEYS[1]) == ) then  
    redis.call('hincrby', KEYS[1], ARGV[2], 1);  
    redis.call('pexpire', KEYS[1], ARGV[1]);  
    return 1;  
end ;  
if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then  
    redis.call('hincrby', KEYS[1], ARGV[2], 1);  
    redis.call('pexpire', KEYS[1], ARGV[1]);  
    return 1;  
end ;  
return ;
```

<span class="bq">加锁代码首先使用 Redis exists  命令判断当前 lock 这个锁是否存在。
 如果锁不存在的话,直接使用 hincrby 创建一个键为 lock  hash 表,并且为 Hash 表中键为 uuid  初始化为 0,然后再次加 1,后再设置过期时间。
 如果当前锁存在,则使用 hexists 判断当前 lock  对应的 hash 表中是否存在 uuid  这个键, 如果存在,再次使用 hincrby  加 1,后再次设置过期时间。
 后如果上述两个逻辑都不符合,直接返回。</span>

**解锁逻辑**

```lua
-- 判断 hash set 可重入 key 的值是否等于   
-- 如果为  代表 该可重入 key 不存在  
if (redis.call('hexists', KEYS[1], ARGV[1]) == ) then  
    return nil;  
end ;  
-- 计算当前可重入次数  
local counter = redis.call('hincrby', KEYS[1], ARGV[1], -1);  
-- 小于等于  代表可以解锁  
if (counter > ) then  
    return ;  
else  
    redis.call('del', KEYS[1]);  
    return 1;  
end ;  
return nil;

```

 <span class="bq">首先使用 hexists  判断 Redis Hash 表是否存给定的域。
 如果 lock 对应 Hash 表不存在,或者 Hash 表不存在 uuid 这个 key,直接返回 nil 。
 若存在的情况下,代表当前锁被其持有,首先使用 hincrby 使可重入次数减 1 ,然后判断计算之后可重入次数,若小于等于 0,则使用 del  删除这把锁。
 解锁代码执行方式与加锁类似,只不过解锁的执行结果返回类型使用 Long 。这里之所以没有跟加锁一样使用 Boolean  ,这是因为解锁 lua 脚本中,三个返回值含义如下:
  1 代表解锁成功,锁被释放
  0 代表可重入次数被减 1 
  null  代表其他线程尝试解锁,解锁失败.</span> <br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redisson 分布式锁<br></span><span class="bq">相关Bean(JavaBean 是特殊的 Java 类,类对象提供任意合法的 Java 数据类型)，以及调用示例。<ul><li> RedissonClient </li><li>RedissonRxClient </li><li>RedissonReactiveClient </li><li>RedisTemplate</li><li> ReactiveRedisTemplate</li></ul> </span> <br>
		-  <span class="bq"><b class="genre1">失败无限重试</b><br> 
```java
RLock lock = redisson.getLock("码哥字节");  
try {  
  
  // 1.常用的种写法  
  lock.lock();  
  
  // 执行业务逻辑  
  .....  
  
} finally {  
  lock.unlock();  
}
```

拿锁失败时会不停的重试,具有 Watch Dog 自动延期机制,默认续 30s 每隔 30/3=10 秒续到30s。</span>

		-  <span class="bq"><b class="genre1">失败超时重试,自动续命</b><br><code> // 尝试拿锁10s后停止重试,获取失败返回false,具有Watch Dog 自动延期机制, 默认续30s 
		boolean flag = lock.tryLock(10, TimeUnit.SECONDS);</code></span> <br>
		-  <span class="bq"><b class="genre1">超时自动释放锁</b><br><code>// 没有Watch Dog ,10s后自动释放,不需要调用 unlock 释放锁。
 lock.lock(10, TimeUnit.SECONDS);</code></span> <br>
		-  <span class="bq"><b class="genre1">超时重试,自动解锁</b><br> 
    ```java
// 尝试加锁,多等待100秒,上锁以后10秒自动解锁,没有 Watch dog 
boolean res = lock.tryLock(100, 10, TimeUnit.SECONDS);  
if (res) {  
   try {  
     ...  
   } finally {  
       lock.unlock();  
   }  
}
    ```
</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Watch Dog 自动延时<br></span><span class="bq">如果获取分布式锁的节点宕机,且这个锁还处于锁定状态,就会出现死锁。
 为了避免这个情况,我们都会给锁设置一个超时自动释放时间。</span> <br>
			-  <span class="bq">假设线程获取锁成功,并设置了 30 s 超时,但是在 30s 内任务还没执行完,锁超时释放了,就会导致其他线程获取不该获取的锁。
 所以,<b class="genre1">Redisson 提供了 watch dog 自动延时机制</b>,提供了一个监控锁的看⻔狗,它的作用是在Redisson 实例被关闭前,不断的延⻓锁的有效期。</span> <br>
			-  <span class="bq">默认情况下,看⻔狗的续期时间是 30s,也可以通过修改 Config.lockWatchdogTimeout  来另行指定。</span> <br>
			-  <span class="bq">Redisson  还提供了可以指定 leaseTime  参数的加锁方法来指定加锁的时间。
 超过这个时间后锁便自动解开了,不会延⻓锁的有效期。
 原理如下图:</span> <br>  ![[E3D9A8FBA6D11E065BFECC2721AE7A59.jpg-2.png]]
			-  <span class="bq">有两个点需要注意: watchDog 只有在未显示指定加锁超时时间(leaseTime)时才会生效。
 lockWatchdogTimeout 设定的时间不要太小 ,比如设置的是 100 毫秒,由于网络直接导致加锁完后,watchdog 去延期时,这个 key 在 redis 中已经被删除了。</span> <br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">源码导读<br></span><span class="bq">在调用 lock 方法时,会终调用到 tryAcquireAsync 。
 调用链为: lock()->tryAcquire->tryAcquireAsync;   </span> #源码阅读不太理解<br>
			-  <span class="bq">scheduleExpirationRenewal  中会调用 renewExpiration  启用了一个 timeout  定时,去执行延期动作。</span> <br>
			-  <span class="bq">scheduleExpirationRenewal  会 调 用 到  renewExpirationAsync , 执 行 下 面 这 段  lua 脚本。
 他主要判断就是 这个锁是否在 redis 中存在,如果存在就进行 pexpire 延期。</span> <br>
 <span><div class="border2"><ul><li>watch dog 在当前节点还存活且任务未完成则每 10 s 给锁续期 30s。</li><li>程序释放锁操作时因为异常没有被执行,那么锁无法被释放,所以释放锁操作一定要放到 finally {} 中; </li><li>要使 watchLog 机制生效 ,lock 时 不要设置 过期时间。</li><li> watchlog 的延时时间 可以由 lockWatchdogTimeout 指定默认延时时间,但是不要设置太小。</li><li>watchdog 会每 lockWatchdogTimeout/3 时间,去延时。</li><li> 通过 lua 脚本实现延迟。</li></ul> </div></span> <br>
 
## Redis分布式锁-可重入锁 - _否极泰来 - 博客园

-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">redis分布式锁-可重入锁 - _否极泰来 - 博客园<br></span><br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">不可重入锁<br></span><span class="bq">即若当前线程执行某个方法已经获取了该锁,那么在方法中尝试再次获取锁时,就会获取不到被阻塞。 同一个人拿一个锁 ,只能拿一次不能同时拿2次。</span> <br>
	-  #Redisson实现可重锁 <span style="color:#a85d45;border-bottom:2px solid #a85d45;">可重入锁<br></span><span class="bq">也叫做递归锁,指的是在同一线程内,外层函数获得锁之后,内层递归函数仍然可以获取到该锁。 说白了就是同一个线程再次进入同样代码时,可以再次拿到该锁。 它的作用是:防止在同一线程中多次获取锁而导致死锁发生。
	-  java的编程中synchronized 和 ReentrantLock都是可重入锁。</span> 
<div class="border2">ReentrantLock的加锁流程是: 1,先判断是否有线程持有锁,没有加锁进行加锁2、如果加锁成功,则设置持有锁的线程是当前线程3、如果有线程持有了锁,则再去判断,是否是当前线程持有了锁4、如果是当前线程持有锁,则加锁数量(state)+1;</div>
```java
final boolean nonfairTryAcquire(int acquires){ 
	final Thread current Thread.currentThread(); 
		int c=getstate();
		 //先判断，c(state)是否等于0，如果等于0，说明没有线程持有锁
		 if(c==0){
		 //通过cas方法把statel的值0替换成1，替换成功说明加锁成功
			 if (compareAndSetstate(0,acquires)){
			  //如果加锁成功，设置持有锁的线程是当前线程
				  setExclusiveownerThread(current);
				   return true; 
				}
			} 
			   else if(current=-getExclusiveOwnerThread()){
			   //判断当前持有锁的线程是否是当前线程
			   //如果是当前线程，则state值加acquires,代表了当前线程加锁了多少次
				   int nextc c acquires; 
				   if (nextc 0)//overflow 
					   throw new Error("Maximum lock count exceeded"); 
				   setstate(nextc); 
				   return true; 
			   }
		   return false;
		}
```
看ReentrantLock的解锁代码我们知道,每次释放锁的时候都对state减1, 当c值等于0的时候,说明锁重入次数也为0了, 最终设置当前持有锁的线程为null,state也设置为0,锁就释放了。;
```java
/** 
	*释放锁
	@param releases 
	@return 
	*/ 
	protected final boolean tryRelease(int releases) {
		int c=getstate()-releases;//state-l减加锁次数
		//如果持有锁的线程，不是当前线程，抛出异常
		if(Thread.currentThread()!=getExclusiveownerThread()) 
			throw new IllegalMonitorStateException(); 
		boolean free = false; 
		if(c == 0){//如果c==0了说明当前线程，已经要释放锁了
			free = true;
			setExclusiveOwnerThread(nul1)://设置当前持有锁的线程为nu11 
		} 
		setState(c);//设置c的值
		return free;
	}
```
<br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">redis要怎么实现可重入的操作<br></span><span class="bq">看ReentrantLock的源码我们知道,它是加锁成功了,记录了当前持有锁的线程,并通过一个int类型的数字,来记录了加锁次数。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">怎么保存当前持有的线程<br></span><span class="bq">redis 的set命令存的是string类型,; 可以保存当前线程的id来解决。
 但是集群环境下我们线程id可能是重复了那怎么解决? 项目在启动的生成一个全局进程id,使用进程id+线程id 那就是唯一的了</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">加锁次数(重入了多少次),怎么记录维护<br></span><span class="bq">Redisson是Redis官方推荐的Java版的Redis客户端。 它基于Java实用工具包中常用接口,为使用者提供了一系列具有分布式特性的常用工具类。 它在网络通信上是基于NIO的Netty框架,保证网络通信的高性能。 在分布式锁的功能上,它提供了一系列的分布式锁;如: 可重入锁(Reentrant Lock)、公平锁(Fair Lock、联锁(MultiLock)、 红锁(RedLock)、 读写锁(ReadWriteLock)等等。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">那么Redisson是如何实现的了</span>
		我们跟一下lock.lock()的代码,发现它最终调用的是  `org.redisson.RedissonLock#tryLockInnerAsync` 的方法;

```java
<T> RFuture<T> tryLockInnerAsync(long waitTime, long leaseTime, TimeUnit unit, long threadId, RedisStrictCommand<T> command) {
		 return evalWriteAsync(getRawName(), LongCodec.INSTANCE, command,
		    "if (redis.call('exists', KEYS[1]) == 0) then " +
		    "redis.call('hincrby', KEYS[1], ARGV[2], 1); " +
		    "redis.call('pexpire', KEYS[1], ARGV[1]); " +
		    "return nil; " +
		    "end; " + 
		    "if (redis.call('hexists', KEYS[1], ARGV[2]) == 1) then " +
		    "redis.call('hincrby', KEYS[1], ARGV[2], 1); " +
		    "redis.call('pexpire', KEYS[1], ARGV[1]); " +
		    "return nil; " +
		    "end; " +
		    "return redis.call('pttl', KEYS[1]);",
		    Collections.singletonList(getRawName()), unit.toMillis(leaseTime), getLockName(threadId));
	}
```

-  exists 查询一个key是否存在
	`EXISTS key [key ...]`  返回值如下的整数结果
		1 如果key存在
		0 如果key不存在
	- `hincrby`  : 将hash中指定域的值增加给定的数字
	- `pexpire` : 设置key的有效时间以毫秒为单位
	- `hexists` : 判断field是否存在于hash中
	- `pttl` : 获取key的有效毫秒数
-  第一个if判断
	- 204行:它是先判断了当前key是否存在,从EXISTS命令我们知道返回值是0说明key不存在,说明没有加锁
	- 205行: `hincrby` 命令是对ARGV[2] = 进程ID+系统ID 进行原子自增加1
	- 206行:是对整个hash设置过期期间
- 8.5、下面来看第二个if判断
	- 209行:判断field是否存在于hash中,如果存在返回1,返回1说明是当前进程+当前线程ID 之前已经获得到锁了
	- 210行:hincrby命令是对 ARGV[2] = 进程ID+系统ID 进行原子自增加1,说明重入次数加1了
	- 211行:再对整个hash设置过期期间; 
- 解锁代码位于 `org.redisson.RedissonLock#unlockInnerAsync`,如下：

``` java
	return evalWriteAsync(getRawName(), LongCodec.INSTANCE, RedisCommands.EVAL_BOOLEAN,
		 "if (redis.call('hexists', KEYS[1], ARGV[3]) == 0) then " +
		 "return nil;" +
		 "end; " +
		 "local counter = redis.call('hincrby', KEYS[1], ARGV[3], -1); " +
		 "if (counter > 0) then " +
		 "redis.call('pexpire', KEYS[1], ARGV[2]); " +
		 "return 0; " +
		 "else " +
		 "redis.call('del', KEYS[1]); " +
		 "redis.call('publish', KEYS[2], ARGV[1]); " +
		 "return 1; " +
		 "end; " +
		 "return nil;",
		 Arrays.asList(getRawName(), getChannelName()), LockPubSub.UNLOCK_MESSAGE, internalLockLeaseTime, getLockName(threadId));
		 }
```

- 阅读完疑问： #阅读完疑问 
	1. Redis没有实现go/C++语言版本的类似的Redssion客户端
	2. Redis客户端是什么？
	3. 我能用c++实现一个吗？可重入锁，原子性？
	4. 可重入锁，为什么redis需要原子性，java不需要？

# 31 事务机制：Redis能实现ACID属性吗？.md

-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">31 事务机制：Redis能实现ACID属性吗？.md<br></span><br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">事务是数据库的一个重要功能<br></span><span class="bq">指对数据进行读写的一系列操作。事务在执行时,会提供专⻔的属性保证,包括原子性(Atomicity)、一致性(Consistency)、隔离性(Isolation)和持久性(Durability),也就是 ACID 属性。这些属性既包括了对事务执行结果的要求,也有对数据库在事务执行前后的数据状态变化的要求。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">原子性<br></span><span class="bq">原子性的要求很明确,就是一个事务中的多个操作必须都完成,或者都不完成。
 业务应用使用事务时,原子性也是最被看重的一个属性。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">一致性<br></span><span class="bq">指数据库中的数据在事务执行前后是一致的。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">隔离性<br></span><span class="bq">要求数据库在执行一个事务时,其它操作无法存取到正在执行事务访问的数据。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">持久性<br></span><span class="bq">数据库执行事务后,数据的修改要被持久化保存下来。当数据库重启后, 数据的值需要是被修改后的值。</span> <br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 如何实现事务?<br></span><span class="bq">事务的执行过程包含三个步骤,Redis 提供了 MULTI、EXEC 两个命令来完成这三个步骤。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第一步<br></span><span class="bq">客户端要使用一个命令显式地表示一个事务的开启。在 Redis 中,这个命令就是 MULTI。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第二步<br></span><span class="bq">客户端把事务中本身要执行的具体操作(例如增删改数据)发送给服务器端。这些操作就是 Redis 本身提供的数据读写命令,例如 GET、SET 等。不过,这些命令虽然被客户端发送到了服务器端,但 Redis 实例只是把这些命令暂存到一个命令队列中,并不会立即执行。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第三步<br></span><span class="bq">客户端向服务器端发送提交事务的命令,让数据库实际执行第二步中发送的具体操作。
 Redis 提供的 EXEC 命令就是执行事务提交的。当服务器端收到 EXEC 命令后,才会实际执行命令队列中的所有命令。</span> <br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">Redis 的事务机制能保证哪些属性?<br></span><span class="bq">Redis 的事务机制能保证哪些属性?; Redis 通过 MULTI、EXEC、DISCARD 和 WATCH 四个命令来支持事务机制,这 4 个命令的作用,我总结在下面的表中,你可以再看下。</span> 
	![[Pasted image 20230710180316.png]]<span class="bq"><b class="focus1">Redis 的事务机制可以保证一致性和隔离性,但是无法保证持久性。</b>不过,因为 Redis 本身是内存数据库,持久性并不是一个必须的属性,我们更加关注的还是原子性、一致性和隔离性这三个属性。
 原子性的情况比较复杂,只有当事务中使用的命令语法有误时,原子性得不到保证,在其它情况下,事务都可以原子性执行。
 所以,我给你一个小建议:严格按照 Redis 的命令规范进行程序开发,并且通过 code review 确保命令的正确性。这样一来,Redis 的事务机制就能被应用在实践中,保证多操作的正确执行。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">原子性<br></span><span class="bq">如果事务正常执行,没有发生任何错误,那么,MULTI 和 EXEC 配合使用,就可以保证多个操作都完成。但是,如果事务执行发生错误了,原子性还能保证吗?我们需要分三种情况来看。
<ul class="genre1"><li>命令入队时就报错,会放弃事务执行,保证原子性; </li><li>命令入队时没报错,实际执行时报错,不保证原子性;</li><li>EXEC 命令执行时实例故障,如果开启了 AOF 日志,可以保证原子性。</li></ul></span> <br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第一种情况是<br></span><span class="bq">在执行 EXEC 命令前,客户端发送的操作命令本身就有错误(比如语法错误,使用了不存在的命令),在命令入队时就被 Redis 实例判断出来了。
 对于这种情况,在命令入队时,Redis 就会报错并且记录下这个错误。此时,我们还能继续提交命令操作。等到执行了 EXEC 命令之后,Redis 就会拒绝执行所有提交的命令操作,返回事务失败的结果。这样一来,事务中的所有命令都不会再被执行了,保证了原子性。</span> <br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第二种情况。<br></span><span class="bq">**事务操作入队时,命令和操作的数据类型不匹配,但 Redis 实例没有检查出错误。**但是,在执行完 EXEC 命令以后,Redis 实际执行这些事务操作时,就会报错。不过,需要注意的是,虽然 Redis 会对错误命令报错,但还是会把正确的命令执行完。在这种情况下,事务的原子性就无法得到保证了。</span> <br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">第三种情况<br></span><span class="bq">在执行事务的 EXEC 命令时,Redis 实例发生了故障,导致事务执行失败。
 在这种情况下,如果 Redis 开启了 AOF 日志,那么,只会有部分的事务操作被记录到 AOF 日志中。我们需要使用 redis-check-aof 工具检查 AOF 日志文件,这个工具可以把未完成的事务操作从AOF 文件中去除。这样一来,我们使用 AOF 恢复实例后,事务操作不会再被执行,从而**保证了原子性**。
 当然,如果 AOF 日志并没有开启,那么实例重启后,数据也都没法恢复了,此时,也就谈不上原子性了。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">一致性<br></span><span class="bq">事务的一致性保证会受到错误命令、实例故障的影响。所以,我们按照命令出错和实例故障的发生时机,分成三种情况来看。</span> <br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">情况一:命令入队时就报错<br></span><span class="bq">在这种情况下,事务本身就会被放弃执行,所以可以保证数据库的一致性。</span> <br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">情况二:命令入队时没报错,实际执行时报错<br></span><span class="bq">在这种情况下,有错误的命令不会被执行,正确的命令可以正常执行,也不会改变数据库的一致性。</span> <br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">情况三:EXEC 命令执行时实例发生故障<br></span><span class="bq">在这种情况下,实例故障后会进行重启,这就和数据恢复的方式有关了,我们要根据实例是否开启了 RDB 或 AOF 来分情况讨论下。
 如果我们没有开启 RDB 或 AOF,那么,实例故障重启后,数据都没有了,数据库是一致的。
 如果我们使用了 RDB 快照,因为 RDB 快照不会在事务执行时执行,所以,事务命令操作的结果不会被保存到 RDB 快照中,使用 RDB 快照进行恢复时,数据库里的数据也是一致的。
 如果我们使用了 AOF 日志,而事务操作还没有被记录到 AOF 日志时,实例就发生了故障,那么, 使用 AOF 日志恢复的数据库数据是一致的。如果只有部分操作被记录到了 AOF 日志,我们可以使用 redis-check-aof 清除事务中已经完成的操作,数据库恢复后也是一致的。
 所以,**总结来说,在命令执行错误或 Redis 发生故障的情况下,Redis 事务机制对一致性属性是有保证的**。</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">隔离性<br></span><span class="bq">事务的隔离性保证,会受到和事务一起执行的并发操作的影响。而事务执行又可以分成命令入队(EXEC 命令执行前)和命令实际执行(EXEC 命令执行后)两个阶段,所以,我们就针对这两个阶段,分成两种情况来分析:<ol start="1"><li> 并发操作在 EXEC 命令前执行,此时,隔离性的保证要使用 WATCH 机制来实现,否则隔离性无法保证; </li><li>并发操作在 EXEC 命令后执行,此时,隔离性可以保证。</li></ol></span> <br>
			-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">WATCH 机制<br></span><span class="bq">在事务执行前,监控一个或多个键的值变化情况,当事务调用 <span class = "high2">EXEC 命令执行时</span> ,WATCH 机制会先检查监控的键是否被其它客户端修改了。如果修改了,就放弃事务执行,避免事务的隔离性被破坏。然后,客户端可以再次执行事务,此时,如果没有并发修改事务数据的操作了,事务就能正常执行,隔离性也得到了保证。
			并发操作在EXEC 命令之后被服务器端接收并执行。
 因为 Redis 是用单线程执行命令,而且,<span class = "high2">EXEC 命令执行后</span> ,Redis 会保证先把命令队列中的所有命令执行完。所以,在这种情况下,并发操作不会破坏事务的隔离性</span> <br>
		-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">持久性<br></span><span class="bq">数据是否持久化保存完全取决于 Redis 的持久化配置模式。
 如果 Redis 没有使用 RDB 或 AOF,那么事务的持久化属性肯定得不到保证。
 如果 Redis 使用了RDB 模式,那么,在一个事务执行后,而下一次的 RDB 快照还未执行前,如果发生了实例宕机, 这种情况下,事务修改的数据也是不能保证持久化的。
 如果 Redis 采用了 AOF 模式,因为 AOF 模式的三种配置选项 no、everysec 和 always 都会存在数据丢失的情况,所以,事务的持久性属性也还是得不到保证。
 所以,不管 Redis 采用什么持久化模式,事务的持久性属性是得不到保证的。</span> <br>
	-  <span style="color:#a85d45;border-bottom:2px solid #a85d45;">数据库回滚机制对比<br></span><span class="bq">传统数据库(例如 MySQL)在执行事务时,会提供回滚机制,当事务执行发生错误时,事务中的所有操作都会撤销,已经修改的数据也会被恢复到事务执行前的状态; Redis 中并没有提供回滚机制。虽然 Redis 提供了 DISCARD 命令,但是,这个命令只能用来主动放弃事务执行,把暂存的命令队列清空,起不到回滚的效果。</span> <br>

