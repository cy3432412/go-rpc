package com.gjerryfe.flink;

import com.alibaba.fastjson.JSONArray;
import com.alibaba.fastjson.JSONObject;
import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.functions.FilterFunction;
import org.apache.flink.api.common.functions.MapFunction;
import org.apache.flink.api.common.serialization.SimpleStringSchema;
import org.apache.flink.api.java.tuple.Tuple2;
import org.apache.flink.api.java.tuple.Tuple7;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.functions.windowing.ProcessWindowFunction;
import org.apache.flink.streaming.api.windowing.assigners.TumblingProcessingTimeWindows;
import org.apache.flink.streaming.api.windowing.time.Time;
import org.apache.flink.streaming.api.windowing.windows.TimeWindow;
import org.apache.flink.streaming.connectors.kafka.FlinkKafkaConsumer;
import org.apache.flink.streaming.connectors.kafka.FlinkKafkaProducer;
import org.apache.flink.util.Collector;

import java.time.Duration;
import java.util.*;
import java.math.RoundingMode;
import java.text.NumberFormat;

//import redis.clients.jedis.Jedis;

public class GroupPowerAnalysis {

	public static void main(String[] args) throws Exception {

		Properties cproperties = new Properties();
		cproperties.setProperty("bootstrap.servers", "10.239.85.238:9092");
		cproperties.setProperty("group.id", "test");
		// set up the streaming execution environment
		final StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();

		FlinkKafkaConsumer<String> myConsumer = new FlinkKafkaConsumer<String>("collectd", new SimpleStringSchema(),
				cproperties);

		myConsumer.assignTimestampsAndWatermarks(WatermarkStrategy.forBoundedOutOfOrderness(Duration.ofSeconds(20)));

		DataStream<String> stream = env.addSource(myConsumer);

		DataStream<Tuple2<String, Tuple7<String, String, String, String, String, Double, Double>>> stream1 = stream.map(
				new MapFunction<String, Tuple2<String, Tuple7<String, String, String, String, String, Double, Double>>>() {
					/**
					 * 
					 */
					private static final long serialVersionUID = 1L;

					public Tuple2<String, Tuple7<String, String, String, String, String, Double, Double>> map(String value) throws Exception {
						String val = value.replace("[", "").replace("]", "");
						JSONObject obj = JSONObject.parseObject(val);
						
						String host = obj.containsKey("host")?obj.getString("host"):"";
						String plug = obj.containsKey("plugin")?obj.getString("plugin"):"";
						String plug_in = obj.containsKey("plugin_instance")?obj.getString("plugin_instance"):"";
						String type = obj.containsKey("type")?obj.getString("type"):"";
						String type_in = obj.containsKey("type_instance")?obj.getString("type_instance"):"";
						double time = obj.containsKey("time")?obj.getDoubleValue("time"):0;
						double values = obj.containsKey("values")?obj.getDoubleValue("values"):0;

						Tuple7<String, String, String, String, String, Double, Double> data;
						data = new Tuple7<>(host, plug, plug_in, type, type_in, time, values);

						Tuple2<String, Tuple7<String, String, String, String, String, Double, Double>> result;
						result = new Tuple2<>(obj.getString("plugin") + "_" + obj.getString("type"), data);
						return result;
					}
				})
				.filter(new FilterFunction<Tuple2<String, Tuple7<String, String, String, String, String, Double, Double>>>() {
					/**
					 * 
					 */
					private static final long serialVersionUID = 1L;

					public boolean filter(
							Tuple2<String, Tuple7<String, String, String, String, String, Double, Double>> value)
							throws Exception {
						return value.f1.f1.equals("pnp") && value.f1.f3.equals("platform_power");
					}
				}

				);
		DataStream<String> convert = stream1.keyBy(f -> f.f0).window(TumblingProcessingTimeWindows.of(Time.seconds(5)))
				.process(new PowerGroupSumProcessWindowFunction());
		convert.print();
		
		Properties pproperties = new Properties();
		pproperties.setProperty("bootstrap.servers", "10.239.85.238:9092");
		pproperties.setProperty("group.id", "testopic");
		
		convert.addSink(new FlinkKafkaProducer<String>("sinkkafka", new SimpleStringSchema(), pproperties)).name("flink2kafka").setParallelism(3);

		env.execute("GroupPowerAnalysis");
	}
}

class PowerGroupSumProcessWindowFunction extends ProcessWindowFunction<Tuple2<String, Tuple7<String, String, String, String, String, Double, Double>>, String, String, TimeWindow> {
	
	/**
	 * 
	 */
	private static final long serialVersionUID = 1L;

	public ArrayList<JSONObject> getredis(){
		/*
		Jedis jd = new Jedis("localhost", 6379);
		Set<String> keys = jd.keys("*");
		Iterator<String> it = keys.iterator();
		while (it.hasNext()) {
			String key = it.next();
			if(jd.type(key).equals("hash")){
				HashMap<String, String> group = new HashMap<>();
				String nodes = jd.hget(key, "nodes").replace("[", "").replace("]", "");
				String power = jd.hget(key, "power_limit");
				group.put("nodes", nodes);
				group.put("power", power);
				all.put(key, group);
			}
		}
		*/
		int power;
		ArrayList<JSONObject> array = new ArrayList<>();
		for(int i=0; i< 2;i++) {
			JSONObject obj = new JSONObject();
			obj.put("group", "rack_"+i);
			List<String> node1 = new ArrayList<>();
			if(i == 0) {
				List<String> namesList = Arrays.asList("demo10", "demo16");
				node1.addAll(namesList);
				power = 200;
			}else {
				node1.add("demo13");
				power = 130;
			}
			obj.put("nodes", node1);
			obj.put("power_provisioning", power);
			array.add(obj);
		}
		return array;
	}
	
	
	@Override
	public void process(String key, Context context,
			Iterable<Tuple2<String, Tuple7<String, String, String, String, String, Double, Double>>> input,
			Collector<String> out) {
		//double format
		NumberFormat nf = NumberFormat.getNumberInstance();
		nf.setRoundingMode(RoundingMode.UP);
		nf.setMaximumFractionDigits(2);
		
		//group(host) data
		Map<String, List<Double>> m = new HashMap<>();
		input.forEach(in->{
			String k = in.f1.f0;
			double v = in.f1.f6;
			if(!m.containsKey(k)) {
				m.put(k, new ArrayList<Double>());
			}
			m.get(k).add(v);
		});
	
		//host average power
		HashMap<String, Double> host_avg = new HashMap<>();
		m.forEach((k, v)->{
			double value = v.stream().mapToDouble(Double::doubleValue).average().getAsDouble();
			host_avg.put(k, Double.valueOf(nf.format(value)));
		});
		//out.collect(host_avg.toString());
		
		//read redis
		ArrayList<JSONObject> rds = getredis();
		//out.collect(rds.toString());
		
		//group power
		List<JSONObject> all = new ArrayList<>();
		rds.forEach(k->{
			JSONObject rack = new JSONObject();
			JSONArray nodes = k.getJSONArray("nodes");
			double group_power = 0;
			ArrayList<HashMap<String, Object>> nodeslist = new ArrayList<>();
			for(Object nd: nodes) {
				HashMap<String, Object> node = new HashMap<>();
				String k1 = nd.toString();
				double avg = host_avg.containsKey(k1)?host_avg.get(k1):0;
				node.put("host", k1);
				node.put("power", avg);
				group_power += avg;
				nodeslist.add(node);
			}
			rack.put("nodes", nodeslist);
			rack.put("group_power", Double.valueOf(nf.format(group_power)));
			rack.put("group", k.getString("group"));
			double provision_power = k.getIntValue("power_provisioning");
			rack.put("power_provisioning", provision_power);
			
			//more than power_provisioning
			if(group_power > provision_power) {
				double average = provision_power / nodeslist.size();
				HashMap<String, Object> power_limit = new HashMap<>();
				HashMap<String, Double> hst = new HashMap<>();
				for(HashMap<String, Object> act: nodeslist) {
					String h = (String) act.get("host");	
					hst.put(h, Double.valueOf(nf.format(average)));
				}
				power_limit.put("power_limit", hst);
				rack.put("actions", power_limit);
			}	
			all.add(rack);
		});
		out.collect(all.toString());
	}
}