import {
  Body,
  Button,
  Container,
  Head,
  Html,
  Link,
  Preview,
  Section,
  Tailwind,
  Text,
} from "@react-email/components";
import * as React from "react";
import { tailwindConfig } from "../tailwind.config";

// Go template placeholders
const NAME = "{{.Name}}";
const RESET_LINK = "{{.ResetLink}}";

export const PasswordReset = () => {
  return (
    <Html>
      <Head />
      <Preview>Redefina sua senha do ZapAgenda</Preview>
      <Tailwind config={tailwindConfig}>
        <Body className="bg-gray-100 font-sans">
          <Container className="bg-white mx-auto my-10 max-w-xl rounded-lg shadow-sm">
            <Section className="px-12 py-8 border-b border-gray-200">
              <Text className="text-2xl font-bold text-brand m-0">
                ZapAgenda
              </Text>
            </Section>

            <Section className="px-12 py-8">
              <Text className="text-xl font-bold text-brand mb-6">
                Redefinir Senha
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-4">
                Olá {NAME},
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-6">
                Recebemos uma solicitação para redefinir a senha da sua conta no
                ZapAgenda. Clique no botão abaixo para criar uma nova senha:
              </Text>

              <Button
                href={RESET_LINK}
                className="bg-brand text-white font-semibold py-3 px-6 rounded-lg"
              >
                Redefinir Senha
              </Button>

              <Text className="text-base text-gray-600 leading-7 mt-6 mb-4">
                Se você não solicitou a redefinição de senha, pode ignorar este
                email com segurança. Sua senha permanecerá inalterada.
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-4">
                Este link expira em 1 hora. Se precisar de um novo link, acesse
                a página de login e clique em "Esqueci minha senha".
              </Text>

              <Text className="text-sm text-gray-400 mt-8">
                Se o botão não funcionar, copie e cole este link no seu
                navegador:
                <br />
                <Link href={RESET_LINK} className="text-brand break-all">
                  {RESET_LINK}
                </Link>
              </Text>
            </Section>

            <Section className="px-12 py-6 border-t border-gray-200">
              <Text className="text-xs text-gray-400 text-center m-0">
                © {new Date().getFullYear()} ZapAgenda. Todos os direitos
                reservados.
              </Text>
            </Section>
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
};

export default PasswordReset;
