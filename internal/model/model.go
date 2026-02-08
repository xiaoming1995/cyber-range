package model

import "time"

// Challenge 挑战题目表 - 存储CTF挑战的基本信息
type Challenge struct {
	ID            string     `gorm:"primaryKey;size:36;comment:题目唯一标识" json:"id"`
	Title         string     `gorm:"size:200;not null;comment:题目标题" json:"title"`
	Description   string     `gorm:"type:text;comment:题目描述(富文本HTML)" json:"description"`
	Hint          string     `gorm:"type:text;comment:题目提示(富文本HTML)" json:"hint,omitempty"`
	Category      string     `gorm:"size:50;comment:题目分类(Web/Pwn/Crypto/Reverse)" json:"category"`
	Difficulty    string     `gorm:"size:20;comment:难度级别(Easy/Medium/Hard)" json:"difficulty"`
	Image         string     `gorm:"size:500;not null;comment:Docker镜像名称(兼容字段)" json:"image"`
	ImageID       string     `gorm:"size:36;index;comment:镜像ID(外键关联docker_images.id)" json:"image_id,omitempty"`
	Port          int        `gorm:"not null;default:80;comment:容器内服务端口" json:"port"`
	MemoryLimit   int64      `gorm:"default:0;comment:内存限制(字节),0表示使用镜像推荐或默认" json:"memory_limit"`
	CPULimit      float64    `gorm:"default:0;comment:CPU限制(核心数),0表示使用镜像推荐或默认" json:"cpu_limit"`
	Privileged    bool       `gorm:"default:false;comment:是否以特权模式运行容器" json:"privileged"`
	Flag          string     `gorm:"size:500;not null;comment:Flag答案(静态模板,不返回给前端)" json:"-"`
	Points        int        `gorm:"not null;default:100;comment:题目分值" json:"points"`
	DockerHostID  string     `gorm:"size:36;index;comment:Docker主机ID(外键关联docker_hosts.id)" json:"docker_host_id,omitempty"`
	Status        string     `gorm:"size:20;default:'unpublished';comment:发布状态(published/unpublished)" json:"status"`
	PublishedAt   *time.Time `gorm:"comment:上架时间" json:"published_at,omitempty"`
	UnpublishedAt *time.Time `gorm:"comment:下架时间" json:"unpublished_at,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// Instance 容器实例表 - 存储用户运行中的靶机实例
type Instance struct {
	ID           string    `gorm:"primaryKey;size:36;comment:实例唯一标识" json:"id"`
	UserID       string    `gorm:"size:36;not null;index:idx_user_challenge;comment:所属用户ID" json:"user_id"`
	ChallengeID  string    `gorm:"size:36;not null;index:idx_user_challenge;comment:关联题目ID" json:"challenge_id"`
	ContainerID  string    `gorm:"size:100;not null;comment:Docker容器ID" json:"container_id"`
	DockerHostID string    `gorm:"size:36;not null;index;comment:Docker主机ID" json:"docker_host_id"`
	Flag         string    `gorm:"size:500;not null;comment:用户专属动态Flag(不返回给前端)" json:"-"`
	Port         int       `gorm:"not null;comment:映射到宿主机的端口号(20000-40000)" json:"port"`
	Status       string    `gorm:"size:20;default:'running';comment:实例状态(running/stopped/expired)" json:"status"`
	ExpiresAt    time.Time `gorm:"not null;index;comment:过期时间(默认1小时后)" json:"expires_at"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
}

// User 用户表 - 存储平台用户信息
type User struct {
	ID           string    `gorm:"primaryKey;size:36;comment:用户唯一标识" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:50;not null;comment:用户名(唯一)" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:100;comment:邮箱地址(唯一)" json:"email"`
	PasswordHash string    `gorm:"size:100;not null;comment:密码哈希值(bcrypt加密)" json:"-"`
	Role         string    `gorm:"size:20;default:'user';comment:用户角色(user/admin)" json:"role"`
	TotalPoints  int       `gorm:"default:0;comment:累计积分" json:"total_points"`
	CreatedAt    time.Time `gorm:"autoCreateTime;comment:注册时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
}

// Submission Flag提交记录表 - 存储所有Flag提交历史
type Submission struct {
	ID          string    `gorm:"primaryKey;size:36;comment:提交记录唯一标识" json:"id"`
	UserID      string    `gorm:"size:36;not null;index;comment:提交用户ID" json:"user_id"`
	ChallengeID string    `gorm:"size:36;not null;index;comment:题目ID" json:"challenge_id"`
	Flag        string    `gorm:"size:500;not null;comment:用户提交的Flag内容" json:"flag"`
	IsCorrect   bool      `gorm:"not null;comment:是否正确(true/false)" json:"is_correct"`
	Points      int       `gorm:"default:0;comment:获得的积分(错误为0)" json:"points"`
	SubmittedAt time.Time `gorm:"autoCreateTime;index;comment:提交时间" json:"submitted_at"`
}

// Admin 管理员表 - 存储后台管理员信息
type Admin struct {
	ID           string     `gorm:"primaryKey;size:36;comment:管理员唯一标识" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:50;not null;comment:管理员用户名(唯一)" json:"username"`
	Email        string     `gorm:"uniqueIndex;size:100;comment:管理员邮箱(唯一)" json:"email"`
	PasswordHash string     `gorm:"size:100;not null;comment:密码哈希值(bcrypt加密)" json:"-"`
	Name         string     `gorm:"size:100;comment:管理员姓名" json:"name"`
	IsActive     bool       `gorm:"default:true;comment:是否激活" json:"is_active"`
	LastLoginAt  *time.Time `gorm:"comment:最后登录时间" json:"last_login_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"-"`
}

// TableName 指定自定义表名（GORM约定）
func (Challenge) TableName() string  { return "challenges" }
func (Instance) TableName() string   { return "instances" }
func (User) TableName() string       { return "users" }
func (Submission) TableName() string { return "submissions" }
func (Admin) TableName() string      { return "admins" }
