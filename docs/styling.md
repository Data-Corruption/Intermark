# Styling

## Title

This is set via the config file `./config.json`

## Logo

The app looks for `logo-darkmode.png` and `logo-lightmode.png` in your content repo's asset directory.

## Footer Special Codes

When editing the text for a footer item, any occurrences of "/year" will be replaced with the current year. (e.g. 2012)

## Adding Fonts

1. Add the font to this repo's css directory `./data/css/`
2. Edit `./data/css/app.css`, for example:

    ```css
    @import "tailwindcss";
    @plugin "@tailwindcss/typography";
    @plugin "daisyui" {
      themes: light, dark --default;
    }

    /*Custom Font Example*/

    @layer base {
      @font-face {
        font-family: 'CustomFont';
        font-style: normal;
        font-weight: 400;
        src: url('./ProggyVector.ttf') format('truetype');
      }
    }

    @layer components {
      .font-custom {
        font-family: 'CustomFont', sans-serif;
      }
    }
    ```

3. Use it in your html.

    ```html
    <span class="font-custom">Hello World</span>
    ```

## Nested Markdown

Any markdown in the root of your file will automatically be converted to html, however anything within html wont. In order to have it be recognized you need to wrap it in a mdsrc tag like so.

```markdown
<div>
<mdsrc>
# This will render properly
Anything inside these "mdsrc" tags will be converted to html along with root level markdown. You can nest like this as deep as you want.
</mdsrc>
# This will not
</div>
```

## Tailwindcss & DaisyUI

For further styling feel free to use Tailwindcss & DaisyUI directly in your markdown or while modifying templates.

## Landing Page

You can set the content for the landing page via assignment in the editor or you can just manually edit it's template at `./data/templates/landing.html`

## Socials

Since these often include icons that need babying with custom html i've opted to leave this up to users entirely. There are two examples by default (discord and github). You can edit the socials at `./data/templates/socials.html`. They are displayed as a list on the bottom of the sidebar.
