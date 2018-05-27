# else branch

# Test to find square root (floor value) of a number

	.data
nStr:		.asciiz "Enter n: "
x:		.word	0
ans:		.word	0
start:		.word	0
end:		.word	0
mid:		.word	0
temp:		.word	0
str:		.asciiz "Square root of n: "

	.text


	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, nStr
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	move	$5, $3		# ans -> $5
	# Store dirty variables back into memory
	sw	$3, x
	sw	$5, ans
	beq	$3, 0, exit

	lw	$3, x
	# Store dirty variables back into memory
	beq	$3, 1, exit

	li	$3, 1		# start -> $3
	sw	$3, start	# spilled start, freed $3
	lw	$3, x
	move	$5, $3		# end -> $5
	# Store dirty variables back into memory
	sw	$5, end

while:
	lw	$3, start
	lw	$5, end
	add	$6, $3, $5	# mid -> $6
	srl	$6, $6, 1	# mid -> $6
	mul	$3, $6, $6	# temp -> $3
	# x is a perfect square
	lw	$5, x
	# Store dirty variables back into memory
	sw	$3, temp
	sw	$6, mid
	beq	$3, $5, perfectSquare

	lw	$3, temp
	lw	$5, x
	# Store dirty variables back into memory
	blt	$3, $5, ifBranch

	lw	$3, mid
	sub	$5, $3, 1	# end -> $5
	lw	$3, start
	# Store dirty variables back into memory
	sw	$5, end
	ble	$3, $5, while

	j	exit

ifBranch:
	lw	$3, mid
	addi	$5, $3, 1	# start -> $5
	move	$6, $3		# ans -> $6
	sw	$6, ans	# spilled ans, freed $6
	lw	$6, end
	# Store dirty variables back into memory
	sw	$5, start
	ble	$5, $6, while

	j	exit

perfectSquare:
	lw	$3, mid
	move	$5, $3		# ans -> $5
	# Store dirty variables back into memory
	sw	$5, ans

exit:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	lw	$3, ans
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	li	$2, 10
	syscall
	.end main
