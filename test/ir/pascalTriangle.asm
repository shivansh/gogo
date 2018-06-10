	.data
twoSpc:		.asciiz "  "
threeSpc:	.asciiz "   "
newline:	.asciiz "\n"
str:		.asciiz "Enter number of rows: "
rows:		.word	0
coef:		.word	0
i:		.word	0
space:		.word	0
temp:		.word	0
k:		.word	0

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime

	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	sw	$3, rows	# spilled rows, freed $3
	li	$3, 1		# coef -> $3
	sw	$3, coef	# spilled coef, freed $3
	li	$3, 0		# i -> $3
	# Store dirty variables back into memory
	sw	$3, i

outerFor:
	lw	$3, i		# i -> $3
	lw	$5, rows	# rows -> $5
	bge	$3, $5, exit

	li	$3, 1		# space -> $3
	# Store dirty variables back into memory
	sw	$3, space

spcFor:
	lw	$3, rows	# rows -> $3
	lw	$5, i		# i -> $5
	sub	$6, $3, $5
	li	$3, 0		# k -> $3
	lw	$5, space	# space -> $5
	# Store dirty variables back into memory
	sw	$3, k
	sw	$6, temp
	bgt	$5, $6, innerFor

	li	$2, 4
	la	$4, twoSpc
	syscall
	lw	$3, space	# space -> $3
	addi	$3, $3, 1
	# Store dirty variables back into memory
	sw	$3, space
	j	spcFor

	li	$3, 0		# k -> $3
	# Store dirty variables back into memory
	sw	$3, k

innerFor:
	lw	$3, k		# k -> $3
	lw	$5, i		# i -> $5
	bgt	$3, $5, endLine

	lw	$3, k		# k -> $3
	beq	$3, 0, labelIf

	lw	$3, i		# i -> $3
	beq	$3, 0, labelIf

	lw	$3, i		# i -> $3
	lw	$5, k		# k -> $5
	sub	$6, $3, $5
	addi	$6, $6, 1
	lw	$3, coef	# coef -> $3
	mul	$3, $3, $6
	div	$3, $3, $5
	# Store dirty variables back into memory
	sw	$3, coef
	sw	$6, temp
	j	labelCoef

labelIf:
	li	$3, 1		# coef -> $3
	# Store dirty variables back into memory
	sw	$3, coef

labelCoef:
	li	$2, 4
	la	$4, threeSpc
	syscall
	li	$2, 1
	lw	$3, coef	# coef -> $3
	move	$4, $3
	syscall
	lw	$3, k		# k -> $3
	addi	$3, $3, 1
	# Store dirty variables back into memory
	sw	$3, k
	j	innerFor

endLine:
	li	$2, 4
	la	$4, newline
	syscall
	lw	$3, i		# i -> $3
	addi	$3, $3, 1
	# Store dirty variables back into memory
	sw	$3, i
	j	outerFor

exit:
	li	$2, 10
	syscall
	.end main
