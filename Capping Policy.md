# Capping Policy

## 1. **Terminology**

+ **Actual Power**

  > 节点实际功耗，可以通过 NM/RAPL/Power Meter获得

+ **Power Capacity**

  > 电源容量，这个值是个预期值，由controller控制

+ **Power Provision**

  > 电源供电量，是一个上限值，绝对不能超过

+ **Default Power**

  > 孩子的默认功率是由其父母通过其优先级计算得出的，所有孩子的默认功率总和等于父母的Power Provision。

+ **Power Pool**

  > 可供分配的能量池

+ **Time Interval**

  > 功率时间间隔

+ **Power Action**

  > 调整动作

+ **Power Action Margin**

  > 功率调整的值

## 2. Power consumption prediction & power pre-allocation

目前还没有办法映射复杂的工作负载和功耗计算集群，工作负载的类型比较稳定，这意味着我们基于Actual Power的时间序列数据进行预测和功率预分配，我们导入2个时间窗口到 做预测：长窗口和短窗口。

在两种情况下，child会尝试向parent申请更多电源：

+ Avg(short-window) > Avg(long-window) + △
+ Actual Power > Power Provision - Power Action Margin

其中△是阻隔噪声的设定值，为此有一个冷静期（在我们的例子中与长时间窗口时间段相同），以避免在短时间内重新触发。

## 3. Power Pool

我们电源管理系统中的任何一个组都有一个电源池来收集其子级未使用的电源，当有子级申请电源操作以获得更多电源预算时，其父级会从电源池中快速分配电源给它。 当Power Pool没有足够的权力时，父母会根据默认power（按优先级计算）来判断是否向其他孩子收回power。

为了确定child是否有未使用的power，我们使用 Max(long window)，对于child，在以下情况下：

+ Max(long window) < Power Provision - Power Action Margin + △



Power from child to parent and the new Power Provision of the node:

+ Power to parent = Power Provision - Power Action Margin -Max(long window) 

+ New Power Provision = Max(long window) + Power Action Margin           

Where △ is a set value to block noise

## 4. **Priority and Default Power allocation**

在我们的电源管理中，当所有孩子都试图从父母那里获得电源而没有足够的电源分配时，这是最坏的情况。 父级将使所有子级在其默认功率下运行，子级 k 的默认功率为

Default power 的公式: 

+ Child Default Powerk = Baselinek + 

  ​      [Parent Power Provision – Sum(Baseline)]×priorityk/Sum(priority)    

其中 BaseLine 和 Priority 由工作负载类型、服务器类型和管理员决定

+ Parent Power Provision = Sum(Child Default Power)                 

所以当一个孩子申请一个Power Action来获得它的默认Power的预算时，它总是可以满足的，如果它的父母的Power Pool没有权力预算，那么父控制器会从Power Provision > Default的其他孩子那里收回power，然后圣化power action

![image-20210714142925767](C:\Users\yangcao1\AppData\Roaming\Typora\typora-user-images\image-20210714142925767.png)