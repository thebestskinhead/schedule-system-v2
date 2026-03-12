SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `student_id` varchar(20) NOT NULL COMMENT '学号/工号',
  `name` varchar(50) NOT NULL COMMENT '姓名',
  `email` varchar(100) NOT NULL,
  `password` varchar(255) NOT NULL COMMENT '加密密码',
  `role` enum('admin','user') DEFAULT 'user' COMMENT '角色',
  `is_active` tinyint DEFAULT '1' COMMENT '是否启用',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `student_id` (`student_id`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_role` (`role`),
  KEY `idx_active` (`is_active`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb3 COMMENT='用户表';

SET FOREIGN_KEY_CHECKS = 1;

INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('1','2505050318','毛岩升','3343503027@qq.com','$2a$10$QVGnt9kK9FQOAnOqEjXhyOspLplqAwcTMK/WVnIaJmx/83vkojb3q','admin','1','2026-02-28 23:51:00','2026-03-02 20:08:38');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('2','2403110229','胡梦玲','3157270485@qq.com','$2a$10$eNjgVtEziAHhRGUUP1bohevzHlnhhtMsk.Vc5kTfNoklkaG7L8Lqi','admin','1','2026-03-01 10:23:56','2026-03-03 10:14:12');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('3','2420010212','钟鹏','3455527297@qq.con','$2a$10$UV2PONLrFZfys8oOKPVjt.eSqaC.fLwhmLvIpb3RHn7Tg0fE7DoAS','admin','1','2026-03-03 09:51:42','2026-03-03 10:14:14');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('4','2512010229','李晓成','1332371365@qq.com','$2a$10$1b9M2pw4Lv67.GhaXjdN8utcw1d7rukSSPTbHtOnixTOIXph3RrOO','user','1','2026-03-03 10:54:51','2026-03-03 10:54:51');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('5','2504010227','廖颐宣','xx76870801@qq.com','$2a$10$e9cO/EAKt3DwhmS7mjKFMOAlPwwImhZED15qm.inTuh2j3QsZGJPG','user','1','2026-03-03 11:08:52','2026-03-03 11:08:52');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('6','2504070410','赵锦耀','whrj666@qq.com','$2a$10$Fo/lGV3p78QGIWbWwxpSj.c3qZ9tOelrZOoIKl8rGDLFZejI4JF/G','user','1','2026-03-03 12:32:54','2026-03-03 12:32:54');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('7','2503030318','傅广骏','3044172061@qq.com','$2a$10$uTAQF3JzaRUZdE538GPvMur/bIiqUb9hwQ3CHyLvkNHOvbhelYlqW','user','1','2026-03-04 12:10:30','2026-03-04 12:10:30');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('8','2515010328','刘语','1705538377@qq.com','$2a$10$NqYQhFmoH1BC1J79oqTgR.PUPzEl5z.F3X0xEFM9vjIDEGc0MdjXe','user','1','2026-03-04 19:12:35','2026-03-04 19:12:35');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('9','2503030130','邓佳琪','2197026723@qq.com','$2a$10$9p0PERNSrcLaB3Do5JRhRu1M5N2M2llyfTbWYsOUSQdrNKL5aozYq','user','1','2026-03-04 23:40:29','2026-03-04 23:40:29');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('10','2509010614','李乐','ll5003@qq.con','$2a$10$M8jhmfe4H96fqKCZc8bU5e3coTF7zB1wvwUX871ANA4y9wYZ45DTO','user','1','2026-03-05 15:11:55','2026-03-05 15:11:55');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('11','2505010606','曾谦易','2539673478@qq.com','$2a$10$zV51/ZUhyS.msCjYlEz3.e2/C3GSzNk.KHWcbDYFKmyI/baCK.rc2','user','1','2026-03-05 22:15:40','2026-03-05 22:15:40');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('12','2505050424','陈慧敏','chm-2008-0211@qq.com','$2a$10$W5cOTq9ZfHeOrMJNd8ezSO9AriT98vBcBSveC7DFtoS5/ERvSEo72','user','1','2026-03-05 22:37:19','2026-03-05 22:37:19');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('13','2501010118','胡兴杰','2793920199@qq.com','$2a$10$s5yiJ0JZG43865XBjuugPOtloRxgRf8jN7WdzIruTAz.ATCEW9NPu','user','1','2026-03-05 22:49:13','2026-03-05 22:49:13');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('14','2505010409','吴乐熙','Honeydear666@qq.com','$2a$10$YsUlH/acfjsLlgPB25vkke6cK7P65pH0KyD3REu5w14tdiLqk9HGq','user','1','2026-03-05 22:59:35','2026-03-05 22:59:35');
INSERT INTO `users` (`id`, `student_id`, `name`, `email`, `password`, `role`, `is_active`, `created_at`, `updated_at`) VALUES ('15','2512010231','拜尔娜•图尔荪','1332371366@qq.com','$2a$10$5kUcm4TkJmfDsZhCICF7Ke.fP52SoEE9xoFSttFn.BtTUEL76Ur2.','user','1','2026-03-11 20:31:20','2026-03-11 20:31:20');

