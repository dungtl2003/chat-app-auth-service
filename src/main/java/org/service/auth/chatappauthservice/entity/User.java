package org.service.auth.chatappauthservice.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import org.service.auth.chatappauthservice.entity.enums.Gender;
import org.service.auth.chatappauthservice.entity.enums.Role;

import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "users")
public class User {

	@Id
	private Long userId;

	@Column(nullable = false, unique = true, columnDefinition = "text")
	private String email;

	@Column(nullable = false, columnDefinition = "text")
	private String username;

	@Column(nullable = false, columnDefinition = "text")
	private String password;

	@Column(nullable = false, columnDefinition = "text")
	private String firstName;

	@Column(nullable = false, columnDefinition = "text")
	private String lastName;

	@Column(nullable = false)
	private LocalDate birthday;

	@Column(nullable = false, columnDefinition = "text")
	private Gender gender;

	@Column(nullable = false, columnDefinition = "text")
	private Role role;

	@Column(nullable = false, columnDefinition = "text")
	private String phoneNumber;

	@Column(columnDefinition = "json")
	private String privacy;

	@Column(nullable = false)
	private Boolean isActive;

	@Column(columnDefinition = "text")
	private String avatarLocation;

	@Column(nullable = false)
	@Builder.Default
	private LocalDateTime createdAt = LocalDateTime.now();

	private LocalDateTime updatedAt;

	private LocalDateTime deletedAt;

}
