package org.service.auth.chatappauthservice.entity;

import com.fasterxml.jackson.annotation.JsonAlias;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.JdbcType;
import org.hibernate.dialect.PostgreSQLEnumJdbcType;
import org.service.auth.chatappauthservice.constants.Gender;
import org.service.auth.chatappauthservice.constants.Role;

import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "user")
public class User implements Cloneable {

	@Id
	@JsonAlias({ "id" })
	@Column(name = "id")
	private String userId;

	@Column(nullable = false, unique = true, columnDefinition = "text")
	private String email;

	@Column(nullable = false, columnDefinition = "text")
	private String username;

	@Column(nullable = false, columnDefinition = "text")
	private String password;

	@Column(name = "first_name", nullable = false, columnDefinition = "text")
	@JsonAlias({ "first_name" })
	private String firstName;

	@Column(name = "last_name", nullable = false, columnDefinition = "text")
	@JsonAlias({ "last_name" })
	private String lastName;

	@Column(nullable = false)
	private LocalDate birthday;

	@Column(nullable = false)
	@Enumerated(EnumType.STRING)
	@JdbcType(PostgreSQLEnumJdbcType.class)
	private Gender gender;

	@Column(nullable = false)
	@Enumerated(EnumType.STRING)
	@JdbcType(PostgreSQLEnumJdbcType.class)
	private Role role;

	@Column(name = "phone_number", nullable = false, columnDefinition = "text")
	@JsonAlias({ "phone_number" })
	private String phoneNumber;

	@Column(nullable = true, columnDefinition = "text")
	private String privacy;

	@Column(name = "is_active", nullable = false)
	@JsonAlias({ "is_active" })
	private Boolean isActive;

	@Column(name = "last_active_at", nullable = false)
	@JsonAlias({ "last_active_at" })
	@Builder.Default
	private LocalDateTime lastActiveAt = LocalDateTime.now();

	@Column(name = "avatar_url", nullable = true, columnDefinition = "text")
	@JsonAlias({ "avatar_url" })
	private String avatarUrl;

	@Column(name = "refresh_tokens", nullable = false, columnDefinition = "text[]")
	@JsonAlias({ "refresh_tokens" })
	@Builder.Default
	private String[] refreshTokens = new String[0];

	@Column(name = "created_at", nullable = false)
	@JsonAlias({ "created_at" })
	@Builder.Default
	private LocalDateTime createdAt = LocalDateTime.now();

	@Column(name = "updated_at", nullable = true)
	@JsonAlias({ "updated_at" })
	private LocalDateTime updatedAt;

	@Column(name = "deleted_at", nullable = true)
	@JsonAlias({ "deleted_at" })
	private LocalDateTime deletedAt;

	@Override
	public User clone() {
		try {
			return (User) super.clone();
		}
		catch (CloneNotSupportedException e) {
			throw new AssertionError();
		}
	}

}
