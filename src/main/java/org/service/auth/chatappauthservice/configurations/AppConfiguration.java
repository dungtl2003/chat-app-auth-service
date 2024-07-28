package org.service.auth.chatappauthservice.configurations;

import lombok.Getter;
import org.service.auth.chatappauthservice.utils.UserDTOMapper;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

@Configuration
@Getter
public class AppConfiguration {

	private final String environment;

	private final long ATLifespanInMs;

	private final long RTLifespanInMs;

	private final int serverPort;

	private final int bcryptStrength;

	private final String apiVersion;

	private final String secretAT;

	private final String secretRT;

	public AppConfiguration() {
		this.environment = System.getenv("ENVIRONMENT") != null ? System.getenv("ENVIRONMENT") : "development";
		this.serverPort = Integer.parseInt(System.getenv("PORT"));
		this.ATLifespanInMs = System.getenv("ACCESS_JWT_LIFESPAN_MS") != null
				? Long.parseLong(System.getenv("ACCESS_JWT_LIFESPAN_MS")) : 10 * 60 * 1000; // 10
		// minutes
		this.RTLifespanInMs = System.getenv("REFRESH_JWT_LIFESPAN_MS") != null
				? Long.parseLong(System.getenv("REFRESH_JWT_LIFESPAN_MS")) : 24 * 60 * 60 * 1000; // 1
		// day
		this.bcryptStrength = System.getenv("BCRYPT_STRENGTH") != null
				? Integer.parseInt(System.getenv("BCRYPT_STRENGTH")) : 12;
		this.apiVersion = System.getenv("API_VERSION");

		this.secretAT = System.getenv("ACCESS_JWT_SECRET");
		this.secretRT = System.getenv("REFRESH_JWT_SECRET");
	}

	@Bean
	public UserDTOMapper userDTOMapper() {
		return new UserDTOMapper();
	}

	@Bean
	public PasswordEncoder passwordEncoder() {
		return new BCryptPasswordEncoder(this.bcryptStrength);
	}

}
