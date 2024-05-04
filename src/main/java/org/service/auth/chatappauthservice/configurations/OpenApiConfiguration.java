package org.service.auth.chatappauthservice.configurations;

import io.swagger.v3.oas.annotations.OpenAPIDefinition;
import io.swagger.v3.oas.annotations.enums.SecuritySchemeIn;
import io.swagger.v3.oas.annotations.enums.SecuritySchemeType;
import io.swagger.v3.oas.annotations.info.Contact;
import io.swagger.v3.oas.annotations.info.Info;
import io.swagger.v3.oas.annotations.info.License;
import io.swagger.v3.oas.annotations.security.SecurityScheme;
import io.swagger.v3.oas.annotations.security.SecuritySchemes;
import io.swagger.v3.oas.annotations.servers.Server;
import org.springdoc.core.models.GroupedOpenApi;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;

@OpenAPIDefinition(
		info = @Info(contact = @Contact(name = "Dung", email = "dungtl2003@gmail.com"),
				description = "Open api documentation for char-app-auth-service", title = "Auth service apis",
				version = "${open-api.version}", license = @License(name = "Api license", url = ""),
				termsOfService = "Terms of service"),
		servers = { @Server(description = "${open-api.server.description}", url = "${open-api.server.url}") })
@SecuritySchemes(value = { @SecurityScheme(name = "authorization",
		description = "Enter the token with the `Bearer ` prefix, e.g. \"Bearer abcde12345\"", scheme = "bearer",
		in = SecuritySchemeIn.HEADER, type = SecuritySchemeType.APIKEY, bearerFormat = "JWT"),
		@SecurityScheme(name = "refresh_token", description = "Enter the refresh_token", in = SecuritySchemeIn.COOKIE,
				type = SecuritySchemeType.APIKEY) })
@Configuration
public class OpenApiConfiguration {

	@Bean
	public GroupedOpenApi groupOpenApi() {
		return GroupedOpenApi.builder()
			.group("auth-service-api")
			.packagesToScan("org.service.auth.chatappauthservice.controller", "org.springframework.boot")
			.build();
	}

}
