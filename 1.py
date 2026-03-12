
class CompactScheduleBitmap:
    """
    紧凑课表位图生成器
    - 30周 × 20位/周 = 600位
    - 每周20位：5天(周一~周五) × 4时段(上午/下午/晚上/深夜)
    - 映射规则：
      bit 0-3: 周一 第1-2节, 第3-4节, 第5-6节, 第7-8节
      bit 4-7: 周二 ...
      bit 8-11: 周三 ...
      bit 12-15: 周四 ...
      bit 16-19: 周五 ...
      bit 20-31: 保留(置0)
    - 1=无课(空闲), 0=有课(占用)
    """
    
    def __init__(self, total_weeks=30):
        self.total_weeks = total_weeks
        self.days = 5  # 周一到周五
        self.blocks_per_day = 4  # 每天4个时段块
        self.bits_per_week = self.days * self.blocks_per_day  # 20位
        
        # 初始化位图：30周 × 20位，初始全1（全部空闲）
        self.bitmap = [0xFFFFF] * total_weeks  # 20个1，即0xFFFFF
        
    def slot_to_block(self, slot_num: int) -> int:
        """将节次(1-10)映射到时段块(0-3)"""
        if 1 <= slot_num <= 2:
            return 0  # 上午1-2节
        elif 3 <= slot_num <= 4:
            return 1  # 上午3-4节
        elif 5 <= slot_num <= 6:
            return 2  # 下午5-6节
        elif 7 <= slot_num <= 10:
            return 3  # 晚上7-8节及以后
        return -1
    
    def mark_occupied(self, week: int, day: int, start_slot: int, end_slot: int):
        """
        标记占用时段为0
        day: 0=周一, 4=周五
        """
        if week < 1 or week > self.total_weeks:
            return
        
        start_block = self.slot_to_block(start_slot)
        end_block = self.slot_to_block(end_slot)
        
        if start_block == -1 or end_block == -1:
            return
        
        # 标记对应的位为0（占用）
        for block in range(start_block, end_block + 1):
            bit_position = day * 4 + block
            if bit_position < self.bits_per_week:
                self.bitmap[week - 1] &= ~(1 << bit_position)  # 将该位清零
    
    def parse_and_mark(self, course_text: str, day: int, default_weeks=None):
        """解析课程文本并标记占用"""
        if default_weeks is None:
            default_weeks = list(range(1, 21))  # 默认1-20周
            
        lines = [l.strip() for l in course_text.split('\n') if l.strip()]
        if len(lines) < 3:
            return
        
        time_info = lines[2]
        
        # 解析周次
        week_match = re.search(r'(\d[\d,\-周单双\(\)]*)\[', time_info)
        if week_match:
            weeks = self.parse_week_pattern(week_match.group(1))
        else:
            weeks = set(default_weeks)
        
        # 解析节次
        slot_match = re.search(r'\[(\d+-\d+)节\]', time_info)
        if slot_match:
            start_slot, end_slot = map(int, slot_match.group(1).split('-'))
        else:
            start_slot, end_slot = 1, 2
        
        # 标记所有周
        for week in weeks:
            self.mark_occupied(week, day, start_slot, end_slot)
    
    def parse_week_pattern(self, pattern: str) -> Set[int]:
        """解析周次（简化版）"""
        weeks = set()
        pattern = pattern.replace('[周]', '').replace('(', '').replace(')', '')
        
        # 处理逗号分隔（如"3,5,7,9"）
        if ',' in pattern and '-' not in pattern.split(',')[0]:
            for w in pattern.split(','):
                try:
                    weeks.add(int(w.strip()))
                except:
                    pass
            return weeks
        
        # 处理范围（如"2-15"）
        match = re.search(r'(\d+)-(\d+)', pattern)
        if match:
            start, end = int(match.group(1)), int(match.group(2))
            
            if '单' in pattern:
                weeks = set(range(start, end + 1, 2))
            elif '双' in pattern:
                start_week = start if start % 2 == 0 else start + 1
                weeks = set(range(start_week, end + 1, 2))
            else:
                weeks = set(range(start, end + 1))
        else:
            try:
                weeks.add(int(pattern))
            except:
                pass
        
        return weeks
    
    def generate_bitmap_array(self) -> List[int]:
        """生成30个32位整数数组"""
        return self.bitmap.copy()
    
    def visualize(self, week: int):
        """可视化某周的位图"""
        val = self.bitmap[week - 1]
        days = ['周一', '周二', '周三', '周四', '周五']
        blocks = ['1-2节', '3-4节', '5-6节', '7-8+节']
        
        print(f"第{week}周位图: 0x{val:05X} (二进制: {val:020b})")
        print("时段分布:")
        for day_idx, day_name in enumerate(days):
            print(f"  {day_name}: ", end="")
            for block_idx in range(4):
                bit_pos = day_idx * 4 + block_idx
                is_free = (val >> bit_pos) & 1
                status = "空闲" if is_free else "占用"
                print(f"[{blocks[block_idx]}]{status} ", end="")
            print()
        print()

# 创建位图实例
bitmap_gen = CompactScheduleBitmap(total_weeks=30)

# 课表数据（与之前相同）
schedule_data = {
    0: [  # 周一
        "高等数学A(2)\n曾石林\n2-15([周])[01-02节]\n第一教学楼205",
        "数据结构与算法设计\n陈轻蕊\n2-17([周])[03-04节]\n第一教学楼405",
        "大学生生涯发展与就业指导\n刘天晶\n15-18([周])[05-06节]\n第一教学楼602",
        "形势与政策\n李育军\n17-18([周])[07-08节]\n第一教学楼602"
    ],
    1: [  # 周二
        "大学外语(2)\n李仙琼\n2-15([周])[01-02节]\n第一教学楼201",
        "离散数学(2)\n李志刚\n2-9([周])[03-04节]\n第一教学楼508",
    ],
    2: [  # 周三
        "数据结构与算法设计\n陈轻蕊\n2-17([周])[01-02节]\n第一教学楼408",
        "高等数学A(2)\n曾石林\n2-14([周])[05-06节]\n第一教学楼308",
        "中国近现代史纲要\n钟声\n2-15([周])[07-08节]\n第一教学楼102"
    ],
    3: [  # 周四
        "线性代数B\n李爱翠\n10-17([周])[01-02节]\n第一教学楼308",
        "离散数学(2)\n李志刚\n2-9([周])[03-04节]\n第一教学楼508",
        "数据结构与算法设计实验\n陈轻蕊\n5-17([周])[05-06节]\n逸夫楼433",
        "大学体育(2)\n李协吉\n3,5,7,9,11,13,15,17([周])[07-08节]\n体育馆"
    ],
    4: [  # 周五
        "高等数学A(2)\n曾石林\n2-14([周])[01-02节]\n第一教学楼208",
        "大学外语(2)\n李仙琼\n2-15([周])[03-04节]\n第一教学楼409",
        "线性代数B\n李爱翠\n10-17([周])[05-06节]\n第一教学楼305"
    ]
}

# 填充数据
for day, courses in schedule_data.items():
    for course_text in courses:
        bitmap_gen.parse_and_mark(course_text, day)

# 输出30个32位整数
print("=" * 60)
print("📊 30周课表位图数据 (1=无课, 0=有课)")
print("格式：每周1个32位整数，低20位有效（5天×4时段）")
print("=" * 60)

bitmap_array = bitmap_gen.generate_bitmap_array()
print("\n十进制数组（可直接复制到代码）：")
print("schedule_bitmap = [")
for i in range(0, 30, 5):
    vals = bitmap_array[i:i+5]
    line = ", ".join(f"0x{val:05X}" for val in vals)
    print(f"    {line},  // 第{i+1}-{i+5}周")
print("]")

print("\n二进制可视化（前10周）：")
for week in range(1, 11):
    val = bitmap_array[week-1]
    # 每4位一组显示
    binary_str = format(val, '020b')
    # 按天分5组
    groups = [binary_str[i:i+4] for i in range(0, 20, 4)]
    formatted = ' '.join(groups)
    print(f"第{week:2d}周: {formatted} (0x{val:05X})")
