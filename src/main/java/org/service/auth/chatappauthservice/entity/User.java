package org.service.auth.chatappauthservice.entity;

import com.fasterxml.jackson.annotation.JsonAlias;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.hibernate.annotations.ColumnTransformer;
import org.service.auth.chatappauthservice.entity.enums.Gender;
import org.service.auth.chatappauthservice.entity.enums.Role;
import org.service.auth.chatappauthservice.utils.PrivacyConverter;

import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "users")
public class User implements Cloneable {

	@Id
	@JsonAlias({ "user_id" })
	private Long userId;

	@Column(nullable = false, unique = true, columnDefinition = "text")
	private String email;

	@Column(nullable = false, columnDefinition = "text")
	private String username;

	@Column(nullable = false, columnDefinition = "text")
	private String password;

	@Column(nullable = false, columnDefinition = "text")
	@JsonAlias({ "first_name" })
	private String firstName;

	@Column(nullable = false, columnDefinition = "text")
	@JsonAlias({ "last_name" })
	private String lastName;

	@Column(nullable = false)
	private LocalDate birthday;

	@Column(nullable = false, columnDefinition = "text")
	private Gender gender;

	@Column(nullable = false, columnDefinition = "text")
	private Role role;

	@Column(nullable = false, columnDefinition = "text")
	@JsonAlias({ "phone_number" })
	private String phoneNumber;

	@Convert(converter = PrivacyConverter.class)
	@Column(columnDefinition = "json")
	@ColumnTransformer(write = "?::json")
	private String privacy;

	@Column(nullable = false)
	@JsonAlias({ "is_active" })
	private Boolean isActive;

	@Column(columnDefinition = "text")
	@JsonAlias({ "avatar_location" })
	private String avatarLocation;

	@Column(columnDefinition = "text[]")
	@JsonAlias({ "refresh_tokens" })
	private String[] refreshTokens;

	@Column(nullable = false)
	@JsonAlias({ "created_at" })
	@Builder.Default
	private LocalDateTime createdAt = LocalDateTime.now();

	@JsonAlias({ "updated_at" })
	private LocalDateTime updatedAt;

	@JsonAlias({ "deleted_at" })
	private LocalDateTime deletedAt;

	@Override
	public User clone() {
		try {
			// TODO: copy mutable state here, so the clone can't change the internals of
			// the original
			return (User) super.clone();
		}
		catch (CloneNotSupportedException e) {
			throw new AssertionError();
		}
	}

}
